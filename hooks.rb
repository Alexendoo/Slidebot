require 'cinch'
require 'date'
require 'json'
require 'net/http'
require 'openssl'
require 'rubygems'
require 'sinatra'
require 'uri'

$config = JSON.parse(File.read('config.json'))

$bot = Cinch::Bot.new do
  configure do |c|
    c.server = $config['irc']['server']
    c.port = $config['irc']['port']
    c.ssl.use = $config['irc']['ssl']['use']
    c.nick = $config['irc']['nick']
    c.user = $config['irc']['user']
    c.password = $config['irc']['password']
    c.channels = $config['irc']['channels']
  end

  on :message, "ping" do |m|
    m.reply "#{m.user.nick}: pong"
  end
end

Thread.new do
  $bot.start
end

def say(msg)
  $config['irc']['channels'].each do |c|
    $bot.Channel(c).send msg
  end
end

post '/' do
  request.body.rewind
  payload_body = request.body.read
  verify_signature(payload_body)
  data = JSON.parse payload_body
  event = request.env['HTTP_X_GITHUB_EVENT']

  begin
    send "receive_#{event}", data
  rescue NameError
    return halt 501, "Not Implemented"
  end
  return halt 200, "Success"
end

def verify_signature(payload_body)
  signature = 'sha1=' + OpenSSL::HMAC.hexdigest(OpenSSL::Digest.new('sha1'), $config['hooks']['secret'], payload_body)
  return halt 500, "Signature mismatch" unless Rack::Utils.secure_compare(signature, request.env['HTTP_X_HUB_SIGNATURE'])
end

def receive_ping(data)
  return halt 200, "pong"
end

def receive_push(d)
  branch = d['ref'].split('/').last
  repo = d['repository']['name']

  puts shorten(d['compare'])

  say "[#{fmt_repo repo}] #{fmt_name d['pusher']['name']} pushed #{fmt_num d['commits'].count} new commit#{d['commits'].count == 1 ? '' : 's'} to #{fmt_branch branch}"
  d['commits'][0..2].each do |commit|
    if commit['message'].include? "\n"
      message = commit['message'].split("\n").first + "..."
    else
      message = commit['message']
    end
    say "#{fmt_repo repo}/#{fmt_branch branch} #{fmt_hash commit['id'][0..6]} #{fmt_name commit['author']['username']} #{message}"
  end
end

def shorten(url)
  uri = URI.parse(url)
  response = Net::HTTP.post_form(URI.parse("http://git.io"), {"url" => uri})

  puts "\n\nCode: #{response.code}\n\nBody: #{response.body}\n\n"
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

def fmt_num(s)
  "\002#{s}\017"
end
