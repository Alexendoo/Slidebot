require 'rubygems'
require 'date'
require 'sinatra'
require 'json'
require 'openssl'

post '/' do
  request.body.rewind
  payload_body = request.body.read
  verify_signature(payload_body)
  data = JSON.parse payload_body
  event = request.env['HTTP_X_GITHUB_EVENT']

  puts "\n#{request.env}"
  puts "\nJSON: #{data.inspect}"

  begin
    send "receive_#{event}", data
  rescue NameError
    return halt 501, "Not Implemented"
  end
  "Sucess"
end

def verify_signature(payload_body)
  signature = 'sha1=' + OpenSSL::HMAC.hexdigest(OpenSSL::Digest.new('sha1'), ENV['SECRET_TOKEN'], payload_body)
  return halt 500, "Signature mismatch" unless Rack::Utils.secure_compare(signature, request.env['HTTP_X_HUB_SIGNATURE'])
end

def receive_push(data)
  data['commits'].sort! do |a,b|
    ta = tb = nil
    begin
      ta = DateTime.parse(a["timestamp"])
    rescue ArgumentError
      ta = Time.at(a["timestamp"].to_i)
    end
    begin
      tb = DateTime.parse(b["timestamp"])
    rescue ArgumentError
      tb = Time.at(b["timestamp"].to_i)
    end
    ta <=> tb
  end

  puts "\nRepo: #{data['repository']['name']}, Pusher: #{data['pusher']['name']}\n"
  data['commits'][0..4].each do |c|
    puts c['message']
  end

  return halt 200
end

def fmt_url(s)
  "\00302\037#{s}\017"
end

def fmt_repo(s)
  "\00313#{s}\017"
end

def fmt_name(s)
  "\00315#{s}\017"
end

def fmt_branch(s)
  "\00306#{s}\017"
end

def fmt_tag(s)
  "\00306#{s}\017"
end

def fmt_hash(s)
  "\00314#{s}\017"
end
