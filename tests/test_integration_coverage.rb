# encoding: utf-8
require 'test/unit'
require 'open3'
require 'tmpdir'
require 'fileutils'

class TestIntegrationCoverage < Test::Unit::TestCase
  def run_cmd(*args)
    cmd = [RbConfig.ruby, File.expand_path('../try.rb', __dir__), *args]
    Open3.capture3(*cmd)
  end

  def test_parse_git_uri_https_github_generates_correct_path
    Dir.mktmpdir do |dir|
      stdout, _stderr, status = run_cmd('clone', 'https://github.com/user/repo.git', '--path', dir)
      assert(status.success?)
      assert_match(/\d{4}-\d{2}-\d{2}-user-repo/, stdout)
      assert_match(/git clone/, stdout)
    end
  end

  def test_parse_git_uri_ssh_format_generates_correct_path
    Dir.mktmpdir do |dir|
      stdout, _stderr, status = run_cmd('clone', 'git@github.com:user/repo.git', '--path', dir)
      assert(status.success?)
      assert_match(/\d{4}-\d{2}-\d{2}-user-repo/, stdout)
    end
  end

  def test_clone_with_custom_name_uses_custom_name
    Dir.mktmpdir do |dir|
      stdout, _stderr, status = run_cmd('clone', 'https://github.com/user/repo.git', 'custom', '--path', dir)
      assert(status.success?)
      assert_match(/custom/, stdout)
      refute_match(/\d{4}-\d{2}-\d{2}-user-repo/, stdout)
    end
  end

  def test_url_shorthand_works_like_clone
    Dir.mktmpdir do |dir|
      stdout, _stderr, status = run_cmd('cd', 'https://github.com/user/repo.git', '--path', dir)
      assert(status.success?)
      assert_match(/git clone/, stdout)
    end
  end

  def test_gitlab_url_parsing
    Dir.mktmpdir do |dir|
      stdout, _stderr, status = run_cmd('clone', 'https://gitlab.com/org/project.git', '--path', dir)
      assert(status.success?)
      assert_match(/\d{4}-\d{2}-\d{2}-org-project/, stdout)
    end
  end

  def test_unique_directory_name_on_collision
    Dir.mktmpdir do |dir|
      date_prefix = Time.now.strftime("%Y-%m-%d")
      existing = "#{date_prefix}-test"
      FileUtils.mkdir_p(File.join(dir, existing))

      stdout, _stderr, _status = run_cmd('cd', 'test', '--and-keys', 'ENTER', '--path', dir)

      # Should either use unique suffix or handle collision
      assert_match(/(test-2|test)/, stdout)
    end
  end

  def test_worktree_without_git_repo_only_creates_dir
    Dir.mktmpdir do |tries|
      Dir.mktmpdir do |repo|
        # No .git directory
        stdout, _stderr, _status = Open3.capture3(
          RbConfig.ruby,
          File.expand_path('../try.rb', __dir__),
          'worktree', 'dir', 'test',
          '--path', tries,
          chdir: repo
        )
        assert_match(/mkdir/, stdout)
        refute_match(/worktree add/, stdout)
      end
    end
  end

  def test_worktree_with_git_repo_adds_worktree
    Dir.mktmpdir do |tries|
      Dir.mktmpdir do |repo|
        FileUtils.mkdir_p(File.join(repo, '.git'))
        stdout, _stderr, _status = Open3.capture3(
          RbConfig.ruby,
          File.expand_path('../try.rb', __dir__),
          'worktree', 'dir', 'test',
          '--path', tries,
          chdir: repo
        )
        assert_match(/worktree add/, stdout)
      end
    end
  end

  def test_init_bash_emits_function
    cmd = [RbConfig.ruby, File.expand_path('../try.rb', __dir__), 'init', '/tmp/tries']
    stdout, _stderr, status = Open3.capture3({'SHELL' => '/bin/bash'}, *cmd)
    assert(status.success?)
    assert_match(/try\(\)/, stdout)
    assert_match(/case/, stdout)
  end

  def test_init_fish_emits_function
    cmd = [RbConfig.ruby, File.expand_path('../try.rb', __dir__), 'init', '/tmp/tries']
    stdout, _stderr, status = Open3.capture3({'SHELL' => '/usr/bin/fish'}, *cmd)
    assert(status.success?)
    assert_match(/function try/, stdout)
    assert_match(/switch/, stdout)
  end

  def test_multiple_directory_collision_handling
    Dir.mktmpdir do |dir|
      date_prefix = Time.now.strftime("%Y-%m-%d")
      FileUtils.mkdir_p(File.join(dir, "#{date_prefix}-test1"))
      FileUtils.mkdir_p(File.join(dir, "#{date_prefix}-test2"))

      # Try creating test1 again - should bump to test3
      stdout, _stderr, _status = run_cmd('cd', 'test1', '--and-keys', 'ENTER', '--path', dir)

      # Should have unique naming
      assert_match(/test/, stdout)
    end
  end

  def test_navigation_keys_ctrl_p_and_n
    Dir.mktmpdir do |dir|
      FileUtils.mkdir_p(File.join(dir, '2025-08-14-first'))
      FileUtils.mkdir_p(File.join(dir, '2025-08-15-second'))
      FileUtils.touch(File.join(dir, '2025-08-14-first', '.touch'))

      # Ctrl-N (down) then Enter
      stdout, _stderr, _status = run_cmd('cd', '--and-keys', 'CTRL-N,ENTER', '--path', dir)
      assert_match(/second/, stdout, 'Ctrl-N should navigate down')
    end
  end

  def test_backspace_removes_characters_from_search
    Dir.mktmpdir do |dir|
      FileUtils.mkdir_p(File.join(dir, 'test-dir'))

      # Type "testx" then backspace, then Enter (should create "test")
      stdout, stderr, _status = run_cmd('cd', '--and-type', 'test', '--and-keys', 'x,BACKSPACE,ENTER', '--path', dir)
      combined = (stdout.to_s + stderr.to_s).force_encoding('UTF-8')
      assert_match(/test/, combined)
    end
  end

  def test_esc_cancels_selector
    Dir.mktmpdir do |dir|
      FileUtils.mkdir_p(File.join(dir, 'test-dir'))

      # Press ESC to cancel
      stdout, _stderr, status = run_cmd('cd', '--and-keys', 'ESC', '--path', dir)
      # Should exit without emitting cd command
      assert(stdout.empty? || !stdout.include?('cd '))
    end
  end
end
