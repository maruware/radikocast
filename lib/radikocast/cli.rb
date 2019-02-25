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
require 'radikocast/s3'

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
      config = load_config
      Radikocast::Schedule.run(config['schedules']) do |program|
        Radikocast.rec(program)
        update_rss(config['podcast']['name'], config['podcast']['host'])
        _publish(config)
      end
    end

    desc 'rss', 'Generate RSS file'
    def rss
      config = load_config
      update_rss(config['podcast']['name'], config['podcast']['host'])
    end

    desc 'publish', 'Publish audio and RSS file'
    def publish
      config = load_config
      _publish(config)
    end

    private
    def load_config
      YAML.safe_load(File.read('config.yml'))
    end

    def update_rss(name, host)
      dst = ENV['DST_DIR']
      xml = RSS.generate(dst, name, host)
      File.write(File.join(dst, 'feed.xml'), xml)
    end

    def _publish(config)
      if config['publish']
        case config['publish']['type']
        when 's3'
          Radikocast.sync_s3(ENV['DST_DIR'], config['publish']['bucket'])
        else
          throw Error.new(`No support publish type for #{config['publish']['type']}`)
        end
      end
    end
  end
end
