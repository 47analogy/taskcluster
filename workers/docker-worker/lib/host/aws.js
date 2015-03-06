/**
Return the appropriate configuration defaults when on aws.
*/

var request = require('superagent-promise');
var debug = require('debug')('docker-worker:configuration:aws');
var log = require('../log')({
  source: 'host/aws'
});

var os = require('os');

function minutes(n) {
  return n * 60;
}

/**
AWS Metadata service endpoint.

@const
@see http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AESDG-chapter-instancedata.html
*/
var BASE_URL = 'http://169.254.169.254/latest';

function* getText(url) {
  var res = yield request.get(url).end();
  // Some meta-data endpoints 404 until they have a value to display (spot node termination)
  var text = res.ok ? res.text : '';
  return text;
}

/**
@return Number Billing cycle interval in seconds.
*/
function* billingCycleInterval() {
  return minutes(60);
}

/**
@return Number of seconds this worker has been running.
*/
function* billingCycleUptime() {
  return os.uptime();
}

/**
Read AWS metadata and user-data to build a configuration for the worker.

@param {String} [baseUrl] optional base url override (for tests).
@return {Object} configuration values.
*/
function* configure (baseUrl) {
  baseUrl = baseUrl || BASE_URL;
  log('configure', { url: BASE_URL });

  // defaults per the metadata
  var config = yield {
    host: getText(baseUrl + '/meta-data/public-hostname'),
    // Since this is aws configuration after all...
    provisionerId: 'aws-provisioner',
    workerId: getText(baseUrl + '/meta-data/instance-id'),
    workerType: getText(baseUrl + '/meta-data/ami-id'),
    workerGroup: getText(baseUrl + '/meta-data/placement/availability-zone'),
    workerNodeType: getText(baseUrl + '/meta-data/instance-type')
  };

  log('metadata', config);

  // AWS Specific shutdown parameters notice this can also be overridden.
  config.shutdown = {
    enabled: true,
    // Always wait 2 minutes minimum prior to shutting down this node.
    minimumCycleSeconds: minutes(2),
  };

  // query the user data for any instance specific overrides set by the
  // provisioner.
  var userdata = yield request.get(baseUrl + '/user-data').
    // Buffer entire response into the .text field of the response.
    buffer(true).
    // Issue the request...
    end();

  if (!userdata.ok || !userdata.text) {
    log('userdata not available')
    return config;
  }
  // parse out overrides from user data
  log('read userdata', { text: userdata.text });
  var overrides = JSON.parse(userdata.text);
  for (var key in overrides) config[key] = overrides[key];

  log('final config', config);
  return config;
};

function* getTerminationTime() {
  var url = BASE_URL + '/meta-data/spot/termination-time';
  var text = yield getText(url);
  return text;
}

module.exports.configure = configure;
module.exports.billingCycleInterval = billingCycleInterval;
module.exports.billingCycleUptime = billingCycleUptime;
module.exports.getTerminationTime = getTerminationTime;
