require 'thor'
require 'zaru'
require 'open3'
require 'rexml/document'
require 'fileutils'
require 'json'
require 'yaml'

require 'radikocast/rss'
require 'radikocast/rec'
require 'radikocast/logger'
require 'radikocast/schedule'
require 'radikocast/program'

module Radikocast
  class CLI < Thor
    desc 'configure NAME, HOST', 'Initialize config'
    def configure(name, host)
      config = {
        'podcast' => {
          'name' => name,
          'host' => host
        },
        'schedule' => []
      }
      File.write('config.yml', YAML.dump(config))
    end

    desc 'rec STATION_ID ITEM_TIMECODE', 'Add new Podcast timefree URL'
    def rec(station, item_timecode)
      program = Radikocast::Program.new(station, item_timecode)
      Radikocast.rec(program)
    end

    desc 'schedule', 'Run scheduler'
    def schedule
      Radikocast::Schedule.run(config['schedules']) do |program|
        # Radikocast.logger.info(program)
        Radikocast.rec(program)
        update_rss
      end
    end

    desc 'rss', 'Generate RSS file'
    def rss
      update_rss
    end

    private
    def config
      YAML.safe_load(File.read('config.yml'))
    end
    def update_rss
      dst = ENV['DST_DIR']
      xml = RSS.generate(dst, config['name'], config['host'])
      File.write(File.join(dst, 'feed.xml'), xml)
    end
  end
end
