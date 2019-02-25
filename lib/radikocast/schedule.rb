require 'rufus-scheduler'

module Radikocast
  module Schedule
    class << self
      def run(schedule_configs)
        schedules = parse_all(schedule_configs)
        Radikocast.logger.debug(schedules)

        scheduler = Rufus::Scheduler.new
        schedules.each do |schedule|
          scheduler.cron schedule[:execution_at] do
            today = Date.today
            yield Radikocast::Program.new(
              schedule[:station],
              gen_start_code(today, schedule[:start][:h], schedule[:start][:m])
            )
          end
        end
        scheduler.join
      end

      def parse_all(schedule_configs)
        schedule_configs.map { |schedule| parse_one(schedule) }
      end

      def parse_one(schedule)
        day_str = schedule['day']
        at_str = schedule['at']
        station = schedule['station']

        # day
        dow = case day_str
              when 'everyday'
                '*'
              when 'weekday'
                '1-5'
              else
                Date::DAYNAMES.map(&:downcase).index(day_str).to_s
              end

        at = parse_at(at_str)

        {
          station: station,
          start: {
            h: at[:start_h],
            m: at[:start_m]
          },
          execution_at: "#{at[:end_m] + 5} #{at[:end_h]} * * #{dow}"
        }
      end

      def parse_at(at_str)
        at_pattern = /^(?<start_h>\d+):(?<start_m>\d+)-(?<end_h>\d+):(?<end_m>\d+)/
        m = at_str.match(at_pattern)
        raise Error, 'Bad at pattern' unless m

        start_h = m['start_h'].to_i
        start_m = m['start_m'].to_i
        end_h = m['end_h'].to_i
        end_m = m['end_m'].to_i

        {
          start_h: start_h,
          start_m: start_m,
          end_h: end_h,
          end_m: end_m
        }
      end

      def gen_start_code(date, start_h, start_m)
        "#{date.strftime('%Y%m%d')}#{format('%02d', start_h)}#{format('%02d', start_m)}00"
      end
    end
  end
end
