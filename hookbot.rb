require 'rubygems'
require 'bundler/setup'

require 'cinch'
require 'json'
require 'openssl'
require 'sinatra'
require 'rest-client'

CONFIG = JSON.parse(File.read('config.json'))

$bot = Cinch::Bot.new do
  configure do |c|
    c.server   = CONFIG['irc']['server']
    c.port     = CONFIG['irc']['port']
    c.ssl.use  = CONFIG['irc']['ssl']['use']
    c.nick     = CONFIG['irc']['nick']
    c.user     = CONFIG['irc']['user']
    c.password = CONFIG['irc']['password']
    c.channels = CONFIG['irc']['channels']
  end

  on :message, /#\d{2,}/ do |m|
    return unless m.user.nick == 'Alexendoo'
    m.message.scan(/#\K\d{2,}/).first(3).each do |issue|
      data = JSON.parse(RestClient.get("https://api.github.com/repos/#{CONFIG['github']['repo']}/issues/#{issue}",
                                       params: { access_token: CONFIG['github']['token'] }))
      m.reply "[#{data['state'] == 'open' ? fmt_good('open') : fmt_bad('closed')}] #{data.key?('pull_request') ? 'Pull request' : 'Issue'} ##{data['number']}: #{data['title']}#{fmt_url shorten data['html_url']}"
    end
  end
end

$bot.loggers.level = :info

Thread.new do
  $bot.start
end

def say(msg)
  CONFIG['irc']['channels'].each do |c|
    $bot.Channel(c).send msg
  end
end

post '/' do
  request.body.rewind
  payload_body = request.body.read
  verify_signature(payload_body) if CONFIG['github'].key?('secret')
  data = JSON.parse payload_body
  event = request.env['HTTP_X_GITHUB_EVENT']

  return halt 202, "Ignored: #{event}" if ignored?(event, data)
  send "receive_#{event}", data
  return halt 200
end

def ignored?(event, data)
  return False unless CONFIG.key?('ignore') && CONFIG['ignore'].key?(event)
  return True if CONFIG['ignore'][event].empty?
  match = (event == 'create' || event == 'delete') ? 'ref_type' : 'action'
  return True if CONFIG['ignore'][event].include? data[match]
  False
end

def verify_signature(payload_body)
  signature = 'sha1=' + OpenSSL::HMAC.hexdigest(OpenSSL::Digest.new('sha1'), CONFIG['github']['secret'], payload_body)
  return halt 500, 'Signature mismatch' unless Rack::Utils.secure_compare(signature, request.env['HTTP_X_HUB_SIGNATURE'])
end

def receive_ping(_d)
  halt 200, 'pong'
end

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
  output =
    case d['deployment_status']['state']
    when 'pending' then 'Pending'
    when 'success' then fmt_good 'Success'
    when 'failure' then fmt_bad 'Failed'
    when 'error' then fmt_bad 'Error'
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
  elsif created > 0 && edited > 0
    output = "edited #{fmt_num edited} and created #{fmt_num created} page#{created == 1 && edited == 1 ? '' : 's'}:#{fallback_url}"
  elsif created > 0
    output = "created #{fmt_num edited} page#{created == 1 ? '' : 's'}:#{fallback_url}"
  else
    output = "edited #{fmt_num created} page#{edited == 1 ? '' : 's'}:#{fallback_url}"
  end
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['sender']['login']} #{output}"
end

def receive_issues(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['issue']['user']['login']} #{d['action']} issue ##{d['issue']['number']}: #{d['issue']['title']}#{fmt_url shorten d['issue']['html_url']}"
end

def receive_issue_comment(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['issue']['user']['login']} commented on issue ##{d['issue']['number']}: #{d['issue']['title']}#{fmt_url shorten d['comment']['html_url']}"
end

def receive_member(_d)
  halt 501
end

def receive_membership(_d)
  halt 501
end

def receive_page_build(_d)
  halt 501
end

def receive_public(_d)
  halt 501
end

def receive_pull_request(d)
  action =
    case d['action']
    when 'synchronize'
      'synchronised'
    when 'closed'
      (d['pull_request']['merged'] == true ? 'merged' : 'closed')
    else
      d['action']
    end
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['pull_request']['user']['login']} #{action} pull request ##{d['pull_request']['number']}: #{d['pull_request']['title']} (#{fmt_branch d['pull_request']['head']['ref']} → #{fmt_branch d['pull_request']['base']['ref']})#{fmt_url shorten d['pull_request']['html_url']}"
end

def receive_pull_request_review_comment(d)
  say "[#{fmt_repo d['repository']['name']}] #{fmt_name d['comment']['user']['login']} commented on pull request ##{d['pull_request']['number']}: #{d['pull_request']['title']}#{fmt_url shorten d['comment']['html_url']}"
end

def receive_push(d)
  branch = d['ref'].split('/').last
  repo = d['repository']['name']
  distinct = d['commits'].select { |commit| commit['distinct'] }

  if distinct.count == 0
    say "[#{fmt_repo repo}] #{fmt_name d['sender']['login']} fast-forwarded #{fmt_repo branch} from #{fmt_hash d['before']} to #{fmt_hash d['after']}:#{fmt_url shorten d['compare']}"
    return halt 200
  end

  say "[#{fmt_repo repo}] #{fmt_name d['sender']['login']} #{d['forced'] ? (fmt_bad 'force pushed') : 'pushed'} #{fmt_num distinct.count} new commit#{distinct.count == 1 ? '' : 's'} to #{fmt_branch branch}:#{fmt_url shorten d['compare']}"
  distinct.first(3).each do |commit|
    message =
      if commit['message'].include? "\n"
        commit['message'].split("\n").first + '...'
      else
        commit['message']
      end
    say "#{fmt_repo repo}/#{fmt_branch branch} #{fmt_hash commit['id']} #{fmt_name commit['author']['username']} #{message}"
  end
end

def receive_repository(_d)
  halt 501
end

def receive_release(_d)
  halt 501
end

def receive_status(_d)
  halt 501
end

def receive_team_add(_d)
  halt 501
end

def receive_watch(_d)
  halt 501
end

def shorten(url)
  (RestClient.post 'https://git.io', url: url).headers[:location]
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
  "\00314#{s.first(7)}\003"
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
