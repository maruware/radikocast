require 'thor'
require 'zaru'
require 'open3'
require 'rexml/document'
require 'fileutils'
require 'json'
require 'yaml'

require 'radikocast/rss'
require 'radikocast/rec'

module Radikocast
  class CLI < Thor
    DST_DIR = File.expand_path('../../../dst', __FILE__)

    desc "configure NAME, HOST", "Initialize config"
    def configure(name, host)
      config = {'name' => name, 'host' => host}
      File.write('config.yml', YAML.dump(config))
    end

    desc "add STATION_ID ITEM_TIMECODE", "Add new Podcast timefree URL"
    def add(station_id, item_timecode)
      Radikocast::rec(station_id, item_timecode)
    end

    desc "rss", "Generate RSS file"
    def rss
      config = YAML.load(File.read('config.yml'))
      xml = RSS.generate(DST_DIR, config['name'], config['host'])
      File.write("#{DST_DIR}/feed.xml", xml)
    end
  end
end
