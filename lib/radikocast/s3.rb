require 'aws-sdk-s3'

module Radikocast
  def self.sync_s3(src_dir, bucket)
    s3 = Aws::S3::Client.new
    Dir.glob(File.join(src_dir, '*.aac')).each do |aac|
      s3.put_object(
        bucket: bucket,
        key: File.basename(aac),
        body: File.open(aac),
        content_type: 'audio/aac',
        acl: 'public-read'
      )
      Radikocast.logger.debug("put s3 #{aac}")
    end
    s3.put_object(
      bucket: bucket,
      key: 'feed.xml',
      body: File.open(File.join(src_dir, 'feed.xml')),
      content_type: 'application/rss+xml',
      acl: 'public-read'
    )
    Radikocast.logger.debug('put s3 feed.xml')
  end
end
