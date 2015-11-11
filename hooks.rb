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

  send "receive_#{event}", data
  return halt 200
end

def verify_signature(payload_body)
  signature = 'sha1=' + OpenSSL::HMAC.hexdigest(OpenSSL::Digest.new('sha1'), $config['secret'], payload_body)
  return halt 500, "Signature mismatch" unless Rack::Utils.secure_compare(signature, request.env['HTTP_X_HUB_SIGNATURE'])
end

def receive_ping(data)
  return halt 200, "pong"
end

#==============================================================================#
# TODO:                                                                        #
# enable/disable in post '/' rather than invidually                            #
# Whitelist/blacklist by other values, not just create: deployment status code #
# Respond to chat #\d\d\d with git.io link + description                       #
#==============================================================================#

def receive_commit_comment(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['comment']['user']['login']} commented on commit #{fmt_hash d['comment']['commit_id']}:#{shorten fmt_url d['comment']['html_url']}"
end

def receive_create(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['sender']['login']} created #{d['ref_type']} #{fmt_tag d['ref']}"
end

def receive_delete(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['sender']['login']} deleted #{d['ref_type']} #{fmt_tag d['ref']}"
end

def receive_deployment(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['deployment']['creator']['login']} deployed #{fmt_hash d['deployment']['sha']} to #{fmt_tag d['deployment']['environment']}"
end

def receive_deployment_status(d)
  case d['deployment_status']['state']
  when 'pending'
    output = 'Pending'
  when 'success'
    output = fmt_good 'Success'
  when 'failure'
    output = fmt_bad 'Failed'
  when 'error'
    output = fmt_bad 'Error'
  end

  say "[#{fmt_repo d['repository']['name']}] Deployment #{fmt_hash d['deployment']['sha']} to #{fmt_tag d['deployment']['environment']}: #{output}"
end

def receive_fork(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['sender']['login']} forked the repository:#{fmt_url shorten d['sender']['html_url']}"
end

def receive_gollum(d)
  edited = 0
  created = 0
  fallback_url = fmt_url shorten "https://github.com/#{d['repository']['full_name']}/wiki/_history"
  d['pages'].each do |page|
    if page['action'] == 'created'
      created += 1
    else
      edited += 1
    end
  end
  if d['pages'].count == 1
    page = d['pages'].first
    output = "#{page['action'] == 'created' ? 'created' : 'edited'} page #{page['title']}:#{fmt_url shorten page['html_url']}"
  elsif created > 0 and edited > 0
    output = "edited #{fmt_num edited} and created #{fmt_num created} page#{created == 1 and edited == 1 ? '' : 's'}:#{fallback_url}"
  elsif created > 0
    output = "created #{fmt_num edited} page#{created == 1 ? '' : 's'}:#{fallback_url}"
  else
    output = "edited #{fmt_num created} page#{edited == 1 ? '' : 's'}:#{fallback_url}"
  end
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['sender']['login']} #{output}"
end

def receive_issues(d)
  return halt 202, "#{d['action']} not enabled" unless $config['hooks']['pull_request']['actions'].include? d['action']
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['issue']['user']['login']} #{d['action']} issue ##{d['issue']['number']}: #{d['issue']['title']}#{fmt_url shorten d['issue']['html_url']}"
end

def receive_issue_comment(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['issue']['user']['login']} commented on issue ##{d['issue']['number']}: #{d['issue']['title']}#{fmt_url shorten d['comment']['html_url']}"
end

def receive_member(d)
  return halt 501
end

def receive_membership(d)
  return halt 501
end

def receive_page_build(d)
  return halt 501
end

def receive_public(d)
  return halt 501
end

def receive_pull_request(d)
  case d['action']
  when "synchronize"
    action = "synchronised"
  when "closed"
    action = (d['pull_request']['merged'] == true ? "merged" : "closed")
  end
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['pull_request']['user']['login']} #{action} pull request ##{d['pull_request']['number']}: #{d['pull_request']['title']} (#{fmt_branch d['pull_request']['head']['ref']} → #{fmt_branch d['pull_request']['base']['ref']})#{fmt_url shorten d['pull_request']['html_url']}"
end

def receive_pull_request_review_comment(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['comment']['user']['login']} commented on pull request ##{d['pull_request']['number']}: #{d['pull_request']['title']}#{fmt_url shorten d['comment']['html_url']}"
end

def receive_push(d)
  branch = d['ref'].split('/').last
  repo = d['repository']['name']
  distinct = d['commits'].select{|commit| commit['distinct']}

  if distinct.count == 0
    say "[#{fmt_repo repo}] #{fmt_name d['sender']['login']} fast-forwarded #{fmt_repo branch} from #{fmt_hash d['before']} to #{fmt_hash d['after']}:#{fmt_url shorten d['compare']}"
    return halt 200
  end

  say "[#{fmt_repo repo}] #{fmt_name d['sender']['login']} #{d['forced'] ? (fmt_bad 'force pushed') : 'pushed'} #{fmt_num distinct.count} new commit#{distinct.count == 1 ? '' : 's'} to #{fmt_branch branch}:#{fmt_url shorten d['compare']}"
  distinct[0..2].each do |commit|
    if commit['message'].include? "\n"
      message = commit['message'].split("\n").first + "..."
    else
      message = commit['message']
    end
    say "#{fmt_repo repo}/#{fmt_branch branch} #{fmt_hash commit['id']} #{fmt_name commit['author']['username']} #{message}"
  end
end

def receive_repository(d)
  return halt 501
end

def receive_release(d)
  return halt 501
end

def receive_status(d)
  return halt 501
end

def receive_team_add(d)
  return halt 501
end

def receive_watch(d)
  return halt 501
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
  "\00314#{s[0..6]}\003"
end

def fmt_num(s)
  "\002#{s}\002"
end

def fmt_good(s)
  "\00303#{s}\003"
end

def fmt_bad(s)
  "\00304#{s}\003"
end
