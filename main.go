package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	docopt "github.com/docopt/docopt-go"
	"github.com/taskcluster/generic-worker/livelog"
	"github.com/taskcluster/httpbackoff"
	"github.com/taskcluster/taskcluster-client-go/queue"
	D "github.com/tj/go-debug"
	"github.com/xeipuuv/gojsonschema"
)

var (
	// Used for logging based on DEBUG environment variable
	// See github.com/tj/go-debug
	debug = D.Debug("generic-worker")
	// General platform independent user settings, such as home directory, username...
	// Platform specific data should be managed in plat_<platform>.go files
	TaskUser OSUser
	// Queue is the object we will use for accessing queue api. See
	// http://docs.taskcluster.net/queue/api-docs/
	Queue *queue.Queue
	// See SignedURLsManager() for more information:
	// signedURsRequestChan is the channel you can pass a channel to, to get
	// back signed urls from the Task Cluster Queue, for querying Azure queues.
	signedURLsRequestChan chan chan *queue.PollTaskUrlsResponse
	// The *currently* one-and-only channel we request signedURLs to be written
	// to. In future we might require more channels to perform requests in
	// parallel, in which case we won't have a single global package var.
	signedURLsResponseChan chan *queue.PollTaskUrlsResponse
	// write to this to close signedurlmanager
	signedDoneChan chan<- bool
	// Channel to request task status updates to the TaskStatusHandler (from
	// any goroutine)
	taskStatusUpdate chan<- TaskStatusUpdate
	// Channel to read errors from after requesting a task status update on
	// taskStatusUpdate channel
	taskStatusUpdateErr <-chan error
	// write to this to close task status update manager
	taskStatusDoneChan chan<- bool
	config             Config

	version = "generic-worker 2.0.0alpha5"
	usage   = `
generic-worker
generic-worker is a taskcluster worker that can run on any platform that supports go (golang).
See http://taskcluster.github.io/generic-worker/ for more details. Essentially, the worker is
the taskcluster component that executes tasks. It requests tasks from the taskcluster queue,
and reports back results to the queue.

  Usage:
    generic-worker run                     [--config         CONFIG-FILE]
                                           [--configure-for-aws]
    generic-worker install                 [--config         CONFIG-FILE]
                                           [--nssm           NSSM-EXE]
                                           [--password       PASSWORD]
                                           [--service-name   SERVICE-NAME]
                                           [--username       USERNAME]
    generic-worker show-payload-schema
    generic-worker --help
    generic-worker --version

  Targets:
    run                                     Runs the generic-worker in an infinite loop.
    show-payload-schema                     Each taskcluster task defines a payload to be
                                            interpreted by the worker that executes it. This
                                            payload is validated against a json schema baked
                                            into the release. This option outputs the json
                                            schema used in this version of the generic
                                            worker.
    install                                 This will install the generic worker as a
                                            Windows service. If the Windows user USERNAME
                                            does not already exist on the system, the user
                                            will be created. This user will be used to run
                                            the service.

  Options:
    --configure-for-aws                     This will create the CONFIG-FILE for an AWS
                                            installation by querying the AWS environment
                                            and setting appropriate values.
    --config CONFIG-FILE                    Json configuration file to use. See
                                            configuration section below to see what this
                                            file should contain.
                                            [default: generic-worker.config]
    --help                                  Display this help text.
    --nssm NSSM-EXE                         The full path to nssm.exe to use for
                                            installing the service.
                                            [default: C:\nssm-2.24\win64\nssm.exe]
    --password PASSWORD                     The password for the username specified
                                            with -u|--username option. If not specified
                                            a random password will be generated.
    --service-name SERVICE-NAME             The name that the Windows service should be
                                            installed under. [default: Generic Worker]
    --username USERNAME                     The Windows user to run the generic worker
                                            Windows service as. If the user does not
                                            already exist on the system, it will be
                                            created. [default: GenericWorker]
    --version                               The release version of the generic-worker.


  Configuring the generic worker:

    The configuration file for the generic worker is specified with -c|--config CONFIG-FILE
    as described above. Its format is a json dictionary of name/value pairs.

        ** REQUIRED ** properties
        =========================

          access_token                      Taskcluster access token used by generic worker
                                            to talk to taskcluster queue.
          client_id                         Taskcluster client id used by generic worker to
                                            talk to taskcluster queue.
          worker_group                      Typically this would be an aws region - an
                                            identifier to uniquely identify which pool of
                                            workers this worker logically belongs to.
          worker_id                         A name to uniquely identify your worker.
          worker_type                       This should match a worker_type managed by the
                                            provisioner you have specified.
          livelog_secret                    This should match the secret used by the
                                            stateless dns server; see
                                            https://github.com/taskcluster/stateless-dns-server
          public_ip                         The IP address for clients to be directed to
                                            for serving live logs; see
                                            https://github.com/taskcluster/livelog and
                                            https://github.com/taskcluster/stateless-dns-server

        ** OPTIONAL ** properties
        =========================

          certificate                       Taskcluster certificate, when using temporary
                                            credentials only.
          provisioner_id                    The taskcluster provisioner which is taking care
                                            of provisioning environments with generic-worker
                                            running on them. [default: aws-provisioner-v1]
          refresh_urls_prematurely_secs     The number of seconds before azure urls expire,
                                            that the generic worker should refresh them.
                                            [default: 310]
          debug                             Logging filter; see
                                            https://github.com/tj/go-debug [default: *]
          livelog_executable                Filepath of LiveLog executable to use; see
                                            https://github.com/taskcluster/livelog
          subdomain                         Subdomain to use in stateless dns name for live
                                            logs; see
                                            https://github.com/taskcluster/stateless-dns-server
                                            [default: taskcluster-worker.net]

    Here is an syntactically valid example configuration file:

            {
              "access_token":               "123bn234bjhgdsjhg234",
              "client_id":                  "hskdjhfasjhdkhdbfoisjd",
              "worker_group":               "dev-test",
              "worker_id":                  "IP_10-134-54-89",
              "worker_type":                "win2008-worker",
              "provisioner_id":             "my-provisioner",
              "livelog_secret":             "baNaNa-SouP4tEa",
              "public_ip":                  "12.24.35.46"
            }


    If an optional config setting is not provided in the json configuration file, the
    default will be taken (defaults documented above).

    If no value can be determined for a required config setting, the generic-worker will
    exit with a failure message.

`
)

// Entry point into the generic worker...
func main() {
	arguments, err := docopt.Parse(usage, nil, true, version, false, true)
	if err != nil {
		fmt.Println("Error parsing command line arguments!")
		panic(err)
	}

	switch {
	case arguments["show-payload-schema"]:
		fmt.Println(taskPayloadSchema())
	case arguments["run"]:
		configureForAws := arguments["--configure-for-aws"].(bool)
		configFile := arguments["--config"].(string)
		config, err = loadConfig(configFile, configureForAws)
		if err != nil {
			fmt.Printf("Error loading configuration from file '%v':\n", configFile)
			fmt.Printf("%v\n", err)
			os.Exit(64)
		}
		runWorker()
		// this returns immediately, as you can runworker in background, so
		// let's wait for a never-arriving message, to avoid exiting program
		forever := make(chan bool)
		<-forever
	case arguments["install"]:
		// platform specific...
		err := install(arguments)
		if err != nil {
			fmt.Println("Error installing generic worker:")
			fmt.Printf("%#v\n", err)
			os.Exit(65)
		}
	}
}

type MissingConfigError struct {
	Setting string
	File    string
}

func (err MissingConfigError) Error() string {
	return "Config setting \"" + err.Setting + "\" must be defined in file \"" + err.File + "\"."
}

func loadConfig(filename string, queryUserData bool) (Config, error) {
	// TODO: would be better to have a json schema, and also define defaults in
	// only one place if possible (defaults also declared in `usage`)

	// first assign defaults
	c := Config{
		Debug:                      "*",
		SubDomain:                  "taskcluster-worker.net",
		ProvisionerId:              "aws-provisioner-v1",
		LiveLogExecutable:          "livelog",
		RefreshUrlsPrematurelySecs: 310,
	}
	// now overlay with values from config file
	configFile, err := os.Open(filename)
	if err != nil {
		return c, err
	}
	defer configFile.Close()
	err = json.NewDecoder(configFile).Decode(&c)
	if err != nil {
		return c, err
	}

	// now overlay with data from amazon, if applicable
	if queryUserData {
		c.updateConfigWithAmazonSettings()
	}

	// now check all values are set
	// TODO: could probably do this with reflection to avoid explicitly listing
	// all members

	fields := []struct {
		value      interface{}
		name       string
		disallowed interface{}
	}{
		{value: c.Debug, name: "debug", disallowed: ""},
		{value: c.ProvisionerId, name: "provisioner_id", disallowed: ""},
		{value: c.RefreshUrlsPrematurelySecs, name: "refresh_urls_prematurely_secs", disallowed: 0},
		{value: c.AccessToken, name: "access_token", disallowed: ""},
		{value: c.ClientId, name: "client_id", disallowed: ""},
		{value: c.WorkerGroup, name: "worker_group", disallowed: ""},
		{value: c.WorkerId, name: "worker_id", disallowed: ""},
		{value: c.WorkerType, name: "worker_type", disallowed: ""},
		{value: c.LiveLogExecutable, name: "livelog_executable", disallowed: ""},
		{value: c.LiveLogSecret, name: "livelog_secret", disallowed: ""},
		{value: c.PublicIP, name: "public_ip", disallowed: net.IP(nil)},
		{value: c.SubDomain, name: "subdomain", disallowed: ""},
	}

	for _, f := range fields {
		if reflect.DeepEqual(f.value, f.disallowed) {
			return c, MissingConfigError{Setting: f.name, File: filename}
		}
	}
	// all config set!
	// now set DEBUG environment variable
	D.Enable(c.Debug)
	return c, nil
}

// returns a channel that you can send 'true' to, to shut it down
func runWorker() chan<- bool {
	// Any custom startup per platform...
	err := startup()
	// any errors are fatal
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	go func() {
		// Queue is the object we will use for accessing queue api
		Queue = queue.New(config.ClientId, config.AccessToken)
		Queue.Certificate = config.Certificate

		// Start the SignedURLsManager in a dedicated go routine, to take care of
		// keeping signed urls up-to-date (i.e. refreshing as old urls expire).
		signedURLsRequestChan, signedURLsResponseChan, signedDoneChan = SignedURLsManager()

		// Start the TaskStatusHandler in a dedicated go routine, to take care of
		// all communication with Queue regarding the status of a TaskRun.
		taskStatusUpdate, taskStatusUpdateErr, taskStatusDoneChan = TaskStatusHandler()

		// loop forever claiming and running tasks!
		for {
			// make sure at least 1 second passes between iterations
			waitASec := time.NewTimer(time.Second * 1)
			taskFound := FindAndRunTask()
			if !taskFound {
				debug("No task claimed from any Azure queue...")
			} else {
				taskCleanup()
			}
			// To avoid hammering queue, make sure there is at least a second
			// between consecutive requests. Note we do this even if a task ran,
			// since a task could complete in less than a second.
			select {
			case <-waitASec.C:
				continue
			case <-done:
				fmt.Println("Shutting down worker...")
				close(done)
				break
			}
		}
		// signedDoneChan <- true
		// taskStatusDoneChan <- true
	}()
	return done
}

// FindAndRunTask loops through the Azure queues in order, to find a task to
// run. If it finds one, it handles all the bookkeeping, as well as running the
// task. Returns true if it successfully claimed a task (regardless of whether
// the task ran successfully) otherwise false.
func FindAndRunTask() bool {
	// Write to the signed urls channel, to request signed urls back on
	// channel c.
	signedURLsRequestChan <- signedURLsResponseChan
	// Read the result.
	signedURLs := <-signedURLsResponseChan
	taskFound := false
	// Each of these signedURLs represent an underlying Azure queue, there
	// are multiple of these so that we can support priority. For this
	// reason the worker must poll the Azure queues in order they are
	// given.
	for _, urlPair := range signedURLs.Queues {
		// try to grab a task using the url pair (url pair = poll url + delete
		// url)
		task, err := SignedURLPair(urlPair).Poll()
		if err != nil {
			// This can be any error at all occurs in queryAzureQueue that
			// prevents us from claiming this task.  Log, and continue.
			debug("%v", err)
			continue
		}
		if task == nil {
			// no task to run, and logging done in function call, so just
			// continue...
			continue
		}
		// Now we found a task, run it, and then exit the loop. This is because
		// the loop is in order of priority, most important first, so we will
		// run the most important task we find, and then return, ignorning
		// remaining urls for lower priority tasks that might still be left to
		// loop through, since by the time we complete the first task, maybe
		// higher priority jobs are waiting, so we need to poll afresh.
		debug("Task found")

		// from this point on we should "break" rather than "continue", since
		// there could be more tasks on the same queue - we only "continue"
		// to next queue if we found nothing on this queue...
		taskFound = true

		// If there is one or more messages the worker must claim the tasks
		// referenced in the messages, and delete the messages.
		taskStatusUpdate <- TaskStatusUpdate{
			Task:   task,
			Status: Claimed,
		}
		err = <-taskStatusUpdateErr
		if err != nil {
			debug("WARN: Not able to claim task %v", task.TaskId)
			debug("%v", err)
			break
		}
		// start the livelogger
		liveLog, err := livelog.New(config.LiveLogExecutable)
		if err != nil {
			debug("FATAL: cannot start livelogger for task " + task.TaskId)
			debug("%v", err)
			os.Exit(88)
		}
		task.liveLog = liveLog
		task.setReclaimTimer()
		task.fetchTaskDefinition()
		err = task.validatePayload()
		if err != nil {
			debug("TASK EXCEPTION: Not able to validate task payload for task %v", task.TaskId)
			debug("%#v", err)
			taskStatusUpdate <- TaskStatusUpdate{
				Task:   task,
				Status: Errored,
				Reason: "malformed-payload", // "invalid-payload"
			}
			reportPossibleError(<-taskStatusUpdateErr)
			break
		}
		err = task.run()
		reportPossibleError(err)
		task.liveLog.Terminate()
		break
	}
	return taskFound
}

func reportPossibleError(err error) {
	if err != nil {
		debug("%v", err)
	}
}

// Queries the given Azure Queue signed url pair (poll url/delete url) and
// translates the Azure response into a Task object
func (urlPair SignedURLPair) Poll() (*TaskRun, error) {
	queueMessagesList := new(QueueMessagesList)
	// To poll an Azure Queue the worker must do a `GET` request to the
	// `signedPollUrl` from the object, representing the Azure queue. To
	// receive multiple messages at once the parameter `&numofmessages=N`
	// may be appended to `signedPollUrl`. The parameter `N` is the
	// maximum number of messages desired, `N` can be up to 32.
	// Since we can only process one task at a time, grab only one.
	resp, _, err := httpbackoff.Get(urlPair.SignedPollUrl + "&numofmessages=1")
	if err != nil {
		debug("%v", err)
		return nil, err
	}
	// When executing a `GET` request to `signedPollUrl` from an Azure queue object,
	// the request will return an XML document on the form:
	//
	// ```xml
	// <QueueMessagesList>
	//     <QueueMessage>
	//       <MessageId>...</MessageId>
	//       <InsertionTime>...</InsertionTime>
	//       <ExpirationTime>...</ExpirationTime>
	//       <PopReceipt>...</PopReceipt>
	//       <TimeNextVisible>...</TimeNextVisible>
	//       <DequeueCount>...</DequeueCount>
	//       <MessageText>...</MessageText>
	//     </QueueMessage>
	//     ...
	// </QueueMessagesList>
	// ```
	// We unmarshal the response into go objects, using the go xml decoder.
	fullBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	reader := strings.NewReader(string(fullBody))
	dec := xml.NewDecoder(reader)
	err = dec.Decode(&queueMessagesList)
	if err != nil {
		debug("ERROR: not able to xml decode the response from the azure Queue:")
		debug(string(fullBody))
		return nil, err
	}
	if len(queueMessagesList.QueueMessages) == 0 {
		debug("Zero tasks returned in Azure XML QueueMessagesList")
		return nil, nil
	}
	if size := len(queueMessagesList.QueueMessages); size > 1 {
		return nil, fmt.Errorf("%v tasks returned in Azure XML QueueMessagesList, even though &numofmessages=1 was specified in poll url", size)
	}

	// at this point we know there is precisely one QueueMessage (== task)
	qm := queueMessagesList.QueueMessages[0]

	// Utility method for replacing a placeholder within a uri with
	// a string value which first must be uri encoded...
	detokeniseUri := func(uri, placeholder, rawValue string) string {
		return strings.Replace(uri, placeholder, strings.Replace(url.QueryEscape(rawValue), "+", "%20", -1), -1)
	}

	// Before using the signedDeleteUrl the worker must replace the placeholder
	// {{messageId}} with the contents of the <MessageId> tag. It is also
	// necessary to replace the placeholder {{popReceipt}} with the URI encoded
	// contents of the <PopReceipt> tag.  Notice, that the worker must URI
	// encode the contents of <PopReceipt> before substituting into the
	// signedDeleteUrl. Otherwise, the worker will experience intermittent
	// failures.

	// Since urlPair is a value, not a pointer, we can update this copy which
	// is associated only with this particular task
	urlPair.SignedDeleteUrl = detokeniseUri(
		detokeniseUri(
			urlPair.SignedDeleteUrl,
			"{{messageId}}",
			qm.MessageId,
		),
		"{{popReceipt}}",
		qm.PopReceipt,
	)

	// Workers should read the value of the `<DequeueCount>` and log messages
	// that alert the operator if a message has been dequeued a significant
	// number of times, for example 15 or more.
	if qm.DequeueCount >= 15 {
		debug("WARN: Queue Message with message id %v has been dequeued %v times!", qm.MessageId, qm.DequeueCount)
		deleteErr := deleteFromAzure(urlPair.SignedDeleteUrl)
		if deleteErr != nil {
			debug("WARN: Not able to call Azure delete URL %v" + urlPair.SignedDeleteUrl)
			debug("%v", deleteErr)
		}
	}

	// To find the task referenced in a message the worker must base64
	// decode and JSON parse the contents of the <MessageText> tag. This
	// would return an object on the form: {taskId, runId}.
	m, err := base64.StdEncoding.DecodeString(qm.MessageText)
	if err != nil {
		// try to delete from Azure, if it fails, nothing we can do about it
		// not very serious - another worker will try to delete it
		debug("ERROR: Not able to base64 decode the Message Text '" + qm.MessageText + "' in Azure QueueMessage response.")
		debug("Deleting from Azure queue as other workers will have the same problem.")
		deleteErr := deleteFromAzure(urlPair.SignedDeleteUrl)
		if deleteErr != nil {
			debug("WARN: Not able to call Azure delete URL %v" + urlPair.SignedDeleteUrl)
			debug("%v", deleteErr)
		}
		return nil, err
	}

	// initialise fields of TaskRun not contained in json string m
	taskRun := TaskRun{
		QueueMessage:  qm,
		SignedURLPair: urlPair,
	}

	// now populate remaining json fields of TaskRun from json string m
	err = json.Unmarshal(m, &taskRun)
	if err != nil {
		debug("Not able to unmarshal json from base64 decoded MessageText '%v'", m)
		debug("%v", err)
		deleteErr := deleteFromAzure(urlPair.SignedDeleteUrl)
		if deleteErr != nil {
			debug("WARN: Not able to call Azure delete URL %v" + urlPair.SignedDeleteUrl)
			debug("%v", deleteErr)
		}
		return nil, err
	}

	return &taskRun, nil
}

// deleteFromAzure will attempt to delete a task from the Azure queue and
// return an error in case of failure
func (task *TaskRun) deleteFromAzure() error {
	if task == nil {
		return fmt.Errorf("Cannot delete task from Azure - task is nil")
	}
	debug("Deleting task " + task.TaskId + " from Azure queue...")
	return deleteFromAzure(task.SignedURLPair.SignedDeleteUrl)
}

// deleteFromAzure is a wrapper around calling an Azure delete URL with error
// handling in case of failure
func deleteFromAzure(deleteUrl string) error {

	// Messages are deleted from the Azure queue with a DELETE request to the
	// signedDeleteUrl from the Azure queue object returned from
	// queue.pollTaskUrls.

	// Also remark that the worker must delete messages if the queue.claimTask
	// operations fails with a 4xx error. A 400 hundred range error implies
	// that the task wasn't created, not scheduled or already claimed, in
	// either case the worker should delete the message as we don't want
	// another worker to receive message later.

	httpCall := func() (*http.Response, error, error) {
		req, err := http.NewRequest("DELETE", deleteUrl, nil)
		if err != nil {
			return nil, nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		return resp, err, nil
	}

	resp, _, err := httpbackoff.Retry(httpCall)

	// Notice, that failure to delete messages from Azure queue is serious, as
	// it wouldn't manifest itself in an immediate bug. Instead if messages
	// repeatedly fails to be deleted, it would result in a lot of unnecessary
	// calls to the queue and the Azure queue. The worker will likely continue
	// to work, as the messages eventually disappears when their deadline is
	// reached. However, the provisioner would over-provision aggressively as
	// it would be unable to tell the number of pending tasks. And the worker
	// would spend a lot of time attempting to claim faulty messages. For these
	// reasons outlined above it's strongly advised that workers logs failures
	// to delete messages from Azure queues.
	if err != nil {
		debug("Not able to delete task from azure queue (delete url: %v)", deleteUrl)
		debug("%v", err)
		return err
	} else {
		debug("Successfully deleted task from azure queue (delete url: %v) with http response code %v.", deleteUrl, resp.StatusCode)
	}
	// no errors occurred, yay!
	return nil
}

func (task *TaskRun) setReclaimTimer() {
	// Reclaiming Tasks
	// ----------------
	// When the worker has claimed a task, it's said to have a claim to a given
	// `taskId`/`runId`. This claim has an expiration, see the `takenUntil`
	// property in the _task status structure_ returned from `queue.claimTask`
	// and `queue.reclaimTask`. A worker must call `queue.reclaimTask` before
	// the claim denoted in `takenUntil` expires. It's recommended that this
	// attempted a few minutes prior to expiration, to allow for clock drift.

	// First time we need to check claim response, after that, need to check reclaim response
	var takenUntil time.Time
	if len(task.TaskReclaimResponse.Status.Runs) > 0 {
		takenUntil = time.Time(task.TaskReclaimResponse.Status.Runs[task.RunId].TakenUntil)
	} else {
		takenUntil = time.Time(task.TaskClaimResponse.Status.Runs[task.RunId].TakenUntil)
	}

	// Attempt to reclaim 3 mins earlier...
	reclaimTime := takenUntil.Add(time.Minute * -3)
	waitTimeUntilReclaim := reclaimTime.Sub(time.Now())
	task.reclaimTimer = time.AfterFunc(
		waitTimeUntilReclaim, func() {
			taskStatusUpdate <- TaskStatusUpdate{
				Task:   task,
				Status: Reclaimed,
			}
			err := <-taskStatusUpdateErr
			if err != nil {
				debug("TASK EXCEPTION due to reclaim failure")
				debug("%v", err)
				taskStatusUpdate <- TaskStatusUpdate{
					Task:   task,
					Status: Errored,
					Reason: "worker-shutdown", // internal error ("reclaim-failed")
				}
				reportPossibleError(<-taskStatusUpdateErr)
				return
			}
			// only set another reclaim timer if the previous reclaim succeeded
			task.setReclaimTimer()
		},
	)
}

func (task *TaskRun) fetchTaskDefinition() {
	// Fetch task definition
	task.Definition = task.TaskClaimResponse.Task
}

func (task *TaskRun) validatePayload() error {
	jsonPayload := task.Definition.Payload
	debug("Json Payload: %s", jsonPayload)
	schemaLoader := gojsonschema.NewStringLoader(taskPayloadSchema())
	docLoader := gojsonschema.NewStringLoader(string(jsonPayload))
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return err
	}
	if result.Valid() {
		debug("The task payload is valid.")
	} else {
		debug("TASK FAIL since the task payload is invalid. See errors:")
		for _, desc := range result.Errors() {
			debug("- %s", desc)
		}
		// Dealing with Invalid Task Payloads
		// ----------------------------------
		// If the task payload is malformed or invalid, keep in mind that the
		// queue doesn't validate the contents of the `task.payload` property,
		// the worker may resolve the current run by reporting an exception.
		// When reporting an exception, using `queue.reportException` the
		// worker should give a `reason`. If the worker is unable execute the
		// task specific payload/code/logic, it should report exception with
		// the reason `malformed-payload`.
		//
		// This can also be used if an external resource that is referenced in
		// a declarative nature doesn't exist. Generally, it should be used if
		// we can be certain that another run of the task will have the same
		// result. This differs from `queue.reportFailed` in the sense that we
		// report a failure if the task specific code failed.
		//
		// Most tasks includes a lot of declarative steps, such as poll a
		// docker image, create cache folder, decrypt encrypted environment
		// variables, set environment variables and etc. Clearly, if decryption
		// of environment variables fail, there is no reason to retry the task.
		// Nor can it be said that the task failed, because the error wasn't
		// cause by execution of Turing complete code.
		//
		// If however, we run some executable code referenced in `task.payload`
		// and the code crashes or exists non-zero, then the task is said to be
		// failed. The difference is whether or not the unexpected behavior
		// happened before or after the execution of task specific Turing
		// complete code.
		taskStatusUpdate <- TaskStatusUpdate{
			Task:   task,
			Status: Errored,
			Reason: "malformed-payload",
		}
		reportPossibleError(<-taskStatusUpdateErr)
		return fmt.Errorf("Validation of payload failed for task %v", task.TaskId)
	}
	return json.Unmarshal(jsonPayload, &task.Payload)
}

func (task *TaskRun) run() error {

	debug("Running task!")
	debug(task.String())

	// Terminating the Worker Early
	// ----------------------------
	// If the worker finds itself having to terminate early, for example a spot
	// nodes that detects pending termination. Or a physical machine ordered to
	// be provisioned for another purpose, the worker should report exception
	// with the reason `worker-shutdown`. Upon such report the queue will
	// resolve the run as exception and create a new run, if the task has
	// additional retries left.
	go func() {
		time.Sleep(time.Second * time.Duration(task.Payload.MaxRunTime))
		taskStatusUpdate <- TaskStatusUpdate{
			Task:   task,
			Status: Aborted,
			// only abort task if it is still running...
			IfStatusIn: map[TaskStatus]bool{Claimed: true, Reclaimed: true},
			Reason:     "malformed-payload", // "max run time (" + strconv.Itoa(task.Payload.MaxRunTime) + "s) exceeded"
		}
		reportPossibleError(<-taskStatusUpdateErr)
	}()

	task.Commands = make([]Command, len(task.Payload.Command))

	// We only report the status at the end of the method, e.g.
	// if a command fails, we still try to upload log files
	// and artifacts. Therefore use these variables to store
	// failure or exception, and at the end of the method
	// report status based on these...
	var finalTaskStatus TaskStatus = Succeeded
	var finalReason string
	var finalError error = nil
	abort := false
	var err error

	for i, _ := range task.Payload.Command {
		task.Commands[i], err = task.generateCommand(i, task.liveLog.LogWriter) // platform specific
		if err != nil {
			debug("%#v", err)
			if finalError == nil {
				debug("TASK EXCEPTION due to not being able to generate command %v", i)
				finalTaskStatus = Errored
				finalReason = "worker-shutdown" // internal error (create-process-error)
				finalError = err
			}
			break
		}
		err = task.Commands[i].osCommand.Start()
		if err != nil {
			debug("%#v", err)
			if finalError == nil {
				debug("TASK EXCEPTION due to not being able to start command %v", i)
				finalTaskStatus = Errored
				finalReason = "worker-shutdown" // internal error (start-process-error)
				finalError = err
			}
			break
		}
		debug("Posting livelog redirect artifact")
		err = task.uploadLiveLog(task.Commands[i].logFile)
		if err != nil {
			debug("%#v", err)
			if finalError == nil {
				debug("TASK EXCEPTION due to problem uploading livelog %v", task.Commands[i].logFile)
				finalTaskStatus = Errored
				finalReason = "worker-shutdown" // internal error (log-upload-failure)
				finalError = err
				// don't break or abort - log upload failure alone shouldn't stop
				// other steps from running
			}
		}
		debug("Waiting for command to finish...")
		// use a different variable for error since we process it later
		err := task.Commands[i].osCommand.Wait()
		// TODO: clean this horrible thing up, and redesign all error handling
		// in this method. It is stinky and terrible.
		errX := task.liveLog.LogWriter.Close()
		if errX != nil {
			panic(errX)
		}

		// Reporting Task Result
		// ---------------------
		// If a task is malformed, the input is invalid, configuration is wrong, or
		// the worker is told to shutdown by AWS before the the task is completed,
		// it should be reported to the queue using `queue.reportException`.
		if err != nil {
			// make sure we abort loop after uploading log file
			debug("%#v", err)
			abort = true
			if finalError == nil {
				// If the task is unsuccessful, ie. exits non-zero, the worker should
				// resolve it using `queue.reportFailed` (this implies test or build
				// failure).
				switch err.(type) {
				case *exec.ExitError:
					finalTaskStatus = Failed
					finalError = err
				default:
					debug("TASK EXCEPTION due to error of type %T when executing command %v", err, i)
					finalTaskStatus = Errored
					finalReason = "worker-shutdown" // internal error (task-crash)
					finalError = err
				}
			}
		}
		err = task.uploadLog(task.Commands[i].logFile)
		if err != nil {
			debug("%#v", err)
			if finalError == nil {
				debug("TASK EXCEPTION due to problem uploading log %v", task.Commands[i].logFile)
				finalTaskStatus = Errored
				finalReason = "worker-shutdown" // internal error (log-upload-failure)
				finalError = err
				// don't break or abort - log upload failure alone shouldn't stop
				// other steps from running
			}
		}
		if abort {
			break
		}
	}

	err = task.postTaskActions()
	if err != nil {
		debug("%#v", err)
		if finalError == nil {
			debug("TASK EXCEPTION when running post-task actions")
			finalTaskStatus = Errored
			finalReason = "worker-shutdown" // internal error (log-concatenation-failure)
			finalError = err
		}
	}

	for _, artifact := range task.PayloadArtifacts() {
		err := task.uploadArtifact(artifact)
		if err != nil {
			debug("%#v", err)
			if finalError == nil {
				switch t := err.(type) {
				case *os.PathError:
					// artifact does not exist or is not readable...
					finalTaskStatus = Failed
					finalError = err
				case httpbackoff.BadHttpResponseCode:
					// if not a 5xx error, then not worth retrying...
					if t.HttpResponseCode/100 != 5 {
						debug("TASK FAIL due to response code %v from Queue when uploading artifact %v", t.HttpResponseCode, artifact)
						finalTaskStatus = Failed
					} else {
						debug("TASK EXCEPTION due to response code %v from Queue when uploading artifact %v", t.HttpResponseCode, artifact)
						finalTaskStatus = Errored
						finalReason = "worker-shutdown" // internal error (upload-failure)
					}
					finalError = err
				default:
					debug("TASK EXCEPTION due to error of type %T", t)
					// could not upload for another reason
					finalTaskStatus = Errored
					finalReason = "worker-shutdown" // internal error (upload-failure)
					finalError = err
				}
			}
		}
	}

	// When the worker has completed the task successfully it should call
	// `queue.reportCompleted`.
	taskStatusUpdate <- TaskStatusUpdate{
		Task:   task,
		Status: finalTaskStatus,
		Reason: finalReason,
	}
	err = <-taskStatusUpdateErr
	if err != nil && finalError == nil {
		debug("%#v", err)
		finalError = err
	}
	return finalError
}

func (task *TaskRun) postTaskActions() error {
	completeLogFile, err := os.Create(filepath.Join(TaskUser.HomeDir, "public", "logs", "all_commands.log"))
	if err != nil {
		return err
	}
	defer completeLogFile.Close()
	for _, command := range task.Commands {
		// unrun commands won't have logFile set...
		if command.logFile != "" {
			debug("Looking for %v", command.logFile)
			commandLog, err := os.Open(filepath.Join(TaskUser.HomeDir, command.logFile))
			if err != nil {
				debug("Not found")
				continue // file does not exist - maybe command did not run
			}
			debug("Found")
			_, err = io.Copy(completeLogFile, commandLog)
			if err != nil {
				debug("Copy failed")
				return err
			}
			debug("Copy succeeded")
			err = commandLog.Close()
			if err != nil {
				return err
			}
		}
	}
	// will only upload if log concatenation succeeded
	return task.uploadLog("public/logs/all_commands.log")
}

func (task *TaskRun) prepEnvVars(cmd *exec.Cmd) {
	workerEnv := os.Environ()
	taskEnv := make([]string, 0)
	for _, j := range workerEnv {
		if !strings.HasPrefix(j, "TASKCLUSTER_ACCESS_TOKEN=") {
			debug("Setting env var: %v", j)
			taskEnv = append(taskEnv, j)
		}
	}
	for i, j := range task.Payload.Env {
		debug("Setting env var: %v=%v", i, j)
		taskEnv = append(taskEnv, i+"="+j)
	}
	cmd.Env = taskEnv
	debug("Environment: %v", taskEnv)
}

// writes to the file configFile with the current generic worker configuration
// (stored in the package variable `config`)
func persistConfig(configFile string) error {
	fmt.Println("Worker ID: " + config.WorkerId)
	fmt.Println("Creating file " + configFile + "...")
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFile, jsonBytes, 0644)
}

func convertNilToEmptyString(val interface{}) string {
	if val == nil {
		return ""
	}
	return val.(string)
}
