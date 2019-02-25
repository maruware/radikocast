module Radikocast
  def self.rec(program)
    cmd = "#{ENV['RADIGO_PATH']} rec -id=#{program.station} -s=#{program.start_code}"
    Radikocast.logger.debug(cmd)
    o, e, s = Open3.capture3(cmd)
    STDERR.puts e if e
    Radikocast.logger.info(o)

    out_lines = o.split("\n")
    dst_line = out_lines[8]
    info = out_lines[5].split('|').map(&:strip)
    title = info[2]

    Zaru.sanitize! title

    dst_dir = ENV['DST_DIR']

    FileUtils.mv(dst_line, dst_dir)
    audio_filename = File.basename(dst_line)

    meta = {
      year: program.year, month: program.month, day: program.day,
      hour: program.hour, minute: program.minute,
      title: title,
      audio_filename: audio_filename,
      audio_size: File.size(File.join(dst_dir, audio_filename))
    }
    meta_filename = File.basename(audio_filename, '.aac') + '.json'
    File.write(File.join(dst_dir, meta_filename), JSON.dump(meta))
  end
end
