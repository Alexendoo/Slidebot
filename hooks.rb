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
    c.modes = $config['irc']['modes']
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
    return halt 501, "No method: #{event}"
  end
  return halt 200
end

def verify_signature(payload_body)
  signature = 'sha1=' + OpenSSL::HMAC.hexdigest(OpenSSL::Digest.new('sha1'), $config['secret'], payload_body)
  return halt 500, "Signature mismatch" unless Rack::Utils.secure_compare(signature, request.env['HTTP_X_HUB_SIGNATURE'])
end

def receive_ping(data)
  return halt 200, "pong"
end

def receive_push(d)
  branch = d['ref'].split('/').last
  repo = d['repository']['name']

  say "[#{fmt_repo repo}] #{fmt_name d['pusher']['name']} pushed #{fmt_num d['commits'].count} new commit#{d['commits'].count == 1 ? '' : 's'} to #{fmt_branch branch}:#{fmt_url shorten d['compare']}"
  d['commits'][0..2].each do |commit|
    if commit['message'].include? "\n"
      message = commit['message'].split("\n").first + "..."
    else
      message = commit['message']
    end
    say "#{fmt_repo repo}/#{fmt_branch branch} #{fmt_hash commit['id'][0..6]} #{fmt_name commit['author']['username']} #{message}"
  end
end

def receive_issues(d)
  return halt 202, "#{d['action']} not enabled" unless $config['hooks']['pull_request']['actions'].include? d['action']
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['issue']['user']['login']} #{d['action']} issue ##{d['issue']['number']}: #{d['issue']['title']}#{fmt_url shorten d['issue']['html_url']}"
end

def receive_issue_comment(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['issue']['user']['login']} commented on issue ##{d['issue']['number']}: #{d['issue']['title']}#{fmt_url shorten d['comment']['html_url']}"
end

def receive_pull_request(d)

end

def shorten(url)
  uri = URI.parse(url)
  response = Net::HTTP.post_form(URI.parse("http://git.io"), {"url" => uri})
  return response['location']
end

def fmt_url(s)
  "\00302 #{s}"
end

def fmt_repo(s)
  "\00313#{s}\003"
end

def fmt_name(s)
  "\00315#{s}\003"
end

def fmt_branch(s)
  "\00306#{s}\003"
end

def fmt_tag(s)
  "\00306#{s}\003"
end

def fmt_hash(s)
  "\00314#{s}\003"
end

def fmt_num(s)
  "\002#{s}\003"
end
