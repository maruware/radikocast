module Radikocast
  class Program
    attr_reader :station, :start_code
    def initialize(station, start_code)
      @station = station
      @start_code = start_code
    end

    def year
      start_code[0..3]
    end
    def month
      start_code[4..5]
    end
    def day
      start_code[6..7]
    end
    def hour
      start_code[8..9]
    end
    def minute
      start_code[10..11]
    end
  end
end
