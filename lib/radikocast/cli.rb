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

module Radikocast
  class CLI < Thor
    desc "configure NAME, HOST", "Initialize config"
    def configure(name, host)
      config = {
        'podcast' => {
          'name' => name,
          'host' => host,
        },
        'schedule' => []
      }
      File.write('config.yml', YAML.dump(config))
    end

    desc "add STATION_ID ITEM_TIMECODE", "Add new Podcast timefree URL"
    def add(station_id, item_timecode)
      Radikocast::rec(station_id, item_timecode)
    end

    desc "rss", "Generate RSS file"
    def rss
      config = YAML.load(File.read('config.yml'))
      dst = ENV['DST_DIR']
      xml = RSS.generate(dst, config['name'], config['host'])
      File.write(File.join(dst, "feed.xml"), xml)
    end
  end
end
