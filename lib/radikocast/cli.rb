require 'thor'
require 'zaru'
require 'open3'
require 'rexml/document'
require 'fileutils'
require 'json'
require 'yaml'

require 'radikocast/rss'
require 'radikocast/scraping'

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
      o, e, s = Open3.capture3("radigo rec -id=#{station_id} -s=#{item_timecode}")
      STDERR.puts e if e
      puts o

      y = item_timecode[0..3]
      m = item_timecode[4..5]
      d = item_timecode[6..7]

      out_lines = o.split("\n")
      info_line = out_lines[5]
      dst_line = out_lines[8]
      info = out_lines[5].split('|').map {|e| e.strip}
      title = info[2]

      Zaru.sanitize! title

      FileUtils.mv(dst_line, "#{DST_DIR}")
      audio_filename = File.basename(dst_line)

      meta = {
        year: y, month: m, day: d,
        title: title,
        audio_filename: audio_filename,
        audio_size: File.size("#{DST_DIR}/#{audio_filename}")
      }
      meta_filename = File.basename(audio_filename, '.aac') + '.json'
      File.write("#{DST_DIR}/#{meta_filename}", JSON.dump(meta))
    end

    desc "rss", "Generate RSS file"
    def rss
      config = YAML.load(File.read('config.yml'))
      xml = RSS.generate(DST_DIR, config['name'], config['host'])
      File.write("#{DST_DIR}/feed.xml", xml)
    end
  end
end
