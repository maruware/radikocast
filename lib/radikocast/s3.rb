require 'aws-sdk-s3'
require 'digest/md5'

module Radikocast
  def self.sync_s3(src_dir, bucket)
    s3 = Aws::S3::Client.new
    res = s3.list_objects_v2({bucket: bucket})
    contents = res.contents

    Dir.glob(File.join(src_dir, '*.aac')).each do |aac|
      sync_object(s3, bucket, aac, 'audio/aac', contents)
    end

    feed_path = File.join(src_dir, 'feed.xml')
    sync_object(s3, bucket, feed_path, 'application/rss+xml; charset=utf-8', contents)
  end

  def self.sync_object(s3, bucket, file_path, content_type, contents)
    key = File.basename(file_path)
    md5 = Digest::MD5.file(file_path).to_s

    content = contents.find {|c| c.key == key && c.etag == "\"#{md5}\""}

    unless content.nil?
      Radikocast.logger.info("[Skip] #{key}")
      return
    end

    s3.put_object(
      bucket: bucket,
      key: key,
      body: File.open(file_path),
      content_type: content_type,
      acl: 'public-read'
    )
    Radikocast.logger.info("[Put] #{key}")
  end
end
