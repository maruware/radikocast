require 'ansi/logger'

module Radikocast
  @logger = ANSI::Logger.new($stdout)

  class << self
    def logger
      @logger
    end
  end
end