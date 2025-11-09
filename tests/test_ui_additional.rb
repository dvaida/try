# encoding: utf-8
require 'test/unit'
require 'stringio'
require_relative '../try.rb'

class TestUIAdditional < Test::Unit::TestCase
  def test_cls_clears_buffers_and_screen
    io = StringIO.new
    UI.puts 'some content'
    UI.cls(io: io)
    out = io.string
    assert_match(/\e\[2J\e\[H/, out, 'cls should emit clear screen and home sequences')
  end

  def test_height_returns_positive_integer
    height = UI.height
    assert(height > 0, 'height should return positive value')
    assert_kind_of(Integer, height)
  end

  def test_width_returns_positive_integer
    width = UI.width
    assert(width > 0, 'width should return positive value')
    assert_kind_of(Integer, width)
  end

  def test_expand_tokens_raises_on_unknown_token
    assert_raises(RuntimeError) do
      UI.expand_tokens('{unknown_token}')
    end
  end

  def test_print_accumulates_to_current_line
    UI.class_variable_set(:@@current_line, "")
    UI.class_variable_set(:@@buffer, [])

    UI.print("hello")
    UI.print(" world")

    current_line = UI.class_variable_get(:@@current_line)
    assert_equal("hello world", current_line)
  end

  def test_puts_adds_to_buffer
    UI.class_variable_set(:@@current_line, "")
    UI.class_variable_set(:@@buffer, [])

    UI.puts("line1")
    UI.puts("line2")

    buffer = UI.class_variable_get(:@@buffer)
    assert_equal(["line1", "line2"], buffer)
  end
end
