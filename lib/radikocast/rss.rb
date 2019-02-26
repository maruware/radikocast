require 'date'
require 'json'
require 'rexml/document'
require 'rexml/formatters/pretty'

module Radikocast
  module RSS
    def generate(dir, title, host, image)
      jsonfiles = Dir.glob(File.join(dir, '*.{json}'))
      metas = jsonfiles.map { |f| JSON.parse(File.read(f)) }

      items = ''
      metas.each do |meta|
        name = "#{meta['year']}/#{meta['month']}/#{meta['day']} - #{meta['title']}"
        url = "#{host}/#{meta['audio_filename']}"
        pub_date = Time.parse("#{meta['year']}-#{meta['month']}-#{meta['day']} #{meta['hour']}:#{meta['minute']}")
        pub_date_str = pub_date.strftime('%a, %d %b %Y %H:%M:%S %z')
        items += <<-XML
          <item>
            <title>#{name}</title>
            <enclosure url="#{url}"
                       length="#{meta['audio_size']}"
                       type="audio/aac" />
            <guid isPermaLink="true">#{url}</guid>
            <pubDate>#{pub_date_str}</pubDate>
          </item>
        XML
      end

      image_elem = if image
                     <<-XML
                      <itunes:image>#{image}</itunes:image>
                     XML
                   else
                     ''
                   end

      xml = <<-XML
      <?xml version="1.0" encoding="utf-8"?>
      <rss xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" version="2.0">
        <channel>
          <title>#{title}</title>
          <language>ja</language>
          #{image_elem}
          #{items}

        </channel>
      </rss>
      XML
      doc = REXML::Document.new(xml)

      pretty_formatter = REXML::Formatters::Pretty.new
      pretty_doc = StringIO.new
      pretty_formatter.write(doc, pretty_doc)
      pretty_doc.string
    end

    module_function :generate
  end
end
