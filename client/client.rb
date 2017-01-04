require 'bugsnag'

Bugsnag.configure do |config|
  config.api_key  = ENV.fetch('BUGSNAG_API_KEY')
  config.endpoint = ENV.fetch('BUGSNAG_ENDPOINT')
  config.use_ssl  = false
end

module App
  def main
    10 / 0
  rescue StandardError => err
    Bugsnag.notify(err, grouping_hash: 'bacon', key: 'value') do |note|
      note.add_tab(:things, a: 1, b: Time.now)
    end
  end
  module_function :main
end # module App

App.main if __FILE__ == $PROGRAM_NAME
