package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/activation"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

// HTTPServer the http server structure.
type HTTPServer struct {
}

// Common strings.
const (
	APIOK  = "ok"
	APIERR = "error"
)

// APIGeneralResp General response to API requests.
type APIGeneralResp struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// JSONResponse Takes a golang structure and converts it to a JSON object for response.
func (s *HTTPServer) JSONResponse(w http.ResponseWriter, resp interface{}) {
	// Encode response as json.
	js, err := json.Marshal(resp)
	if err != nil {
		// Error should not happen normally...
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// If no err, we can set content type header and send response.
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	w.Write([]byte{'\n'})
}

// APISendGeneralResp Send a standard response.
func (s *HTTPServer) APISendGeneralResp(w http.ResponseWriter, status, err string) {
	resp := APIGeneralResp{}
	resp.Status = status
	resp.Error = err
	s.JSONResponse(w, resp)
}

// registerHandlers HTTP server handlers.
func (s *HTTPServer) registerHandlers(r *http.ServeMux) {
	// For this project, we only handle requests to /.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Parse form data.
		err := r.ParseMultipartForm(32 << 20)
		if err == http.ErrNotMultipart {
			err = r.ParseForm()
		}
		if err != nil {
			fmt.Println(err)
			s.APISendGeneralResp(w, APIERR, "Bad request")
			return
		}

		// Verify we are authorized.
		if r.Form.Get("token") != app.config.APIToken {
			s.APISendGeneralResp(w, APIERR, "Unauthorized")
			return
		}

		// Get call details.
		channel := r.Form.Get("channel")
		if channel == "" {
			channel = app.config.DefaultChannel
		}

		callerId := r.Form.Get("caller_id")
		if callerId == "" {
			callerId = app.config.DefaultCallerId
		}

		waitTime, err := strconv.ParseUint(r.Form.Get("wait_time"), 10, 64)
		if err != nil {
			waitTime = app.config.DefaultWaitTime
		}

		maxRetries, err := strconv.ParseUint(r.Form.Get("max_retries"), 10, 64)
		if err != nil {
			maxRetries = app.config.DefaultMaxRetries
		}

		retryTime, err := strconv.ParseUint(r.Form.Get("retry_time"), 10, 64)
		if err != nil {
			retryTime = app.config.DefaultRetryTime
		}

		account := r.Form.Get("account")
		if account == "" {
			account = app.config.DefaultCallerId
		}

		application := r.Form.Get("application")
		if application == "" || app.config.PreventAPIApplication {
			application = app.config.DefaultApplication
		}

		data := r.Form.Get("data")
		if data == "" || app.config.PreventAPIApplication {
			data = app.config.DefaultData
		}

		context := r.Form.Get("context")
		if context == "" {
			context = app.config.DefaultContext
		}

		extension := r.Form.Get("extension")
		if context == "" {
			extension = app.config.DefaultExtension
		}

		priority := r.Form.Get("priority")
		if context == "" {
			priority = app.config.DefaultPriority
		}

		setVar := make(map[string]string)
		parsed, err := url.ParseQuery(r.Form.Get("set_var"))
		if err != nil {
			setVar = app.config.DefaultSetVar
		} else {
			for key, value := range parsed {
				setVar[key] = value[0]
			}
		}

		archiveVal := strings.ToLower(r.Form.Get("archive"))
		archive := false
		if archiveVal == "true" || archiveVal == "yes" {
			archive = true
		} else if archiveVal != "false" && archiveVal != "no" {
			archive = app.config.DefaultArchive
		}

		schedule := r.Form.Get("schedule")

		if channel == "" || (application == "" && context == "") {
			s.APISendGeneralResp(w, APIERR, "Required options not set")
			return
		}

		// Setup call file details.
		outgoingCallName := "outgoing-call-" + strconv.Itoa(rand.Int())
		spoolFileName := path.Join(app.config.AsteriskSpoolDir, outgoingCallName)
		outgoingFileName := path.Join(app.config.AsteriskSpoolDir, "outgoing", outgoingCallName)

		callFile, err := os.Create(spoolFileName)
		if err != nil {
			fmt.Println(err)
			s.APISendGeneralResp(w, APIERR, "Unable to create call file")
			return
		}

		// Write call details.
		callFile.WriteString("Channel: " + channel + "\n")

		if callerId != "" {
			callFile.WriteString("Callerid: " + callerId + "\n")
		}

		if waitTime != 0 {
			callFile.WriteString("WaitTime: " + strconv.FormatUint(waitTime, 10) + "\n")
		}

		if maxRetries != 0 {
			callFile.WriteString("MaxRetries: " + strconv.FormatUint(maxRetries, 10) + "\n")
		}

		if retryTime != 0 {
			callFile.WriteString("RetryTime: " + strconv.FormatUint(retryTime, 10) + "\n")
		}

		if account != "" {
			callFile.WriteString("Account: " + account + "\n")
		}

		if application != "" {
			callFile.WriteString("Application: " + application + "\n")
		}

		if data != "" {
			callFile.WriteString("Data: " + data + "\n")
		}

		if context != "" {
			callFile.WriteString("Context: " + context + "\n")
		}

		if extension != "" {
			callFile.WriteString("Extension: " + extension + "\n")
		}

		if priority != "" {
			callFile.WriteString("Priority: " + priority + "\n")
		}

		for key, value := range setVar {
			callFile.WriteString("Setvar: " + key + "=" + value + "\n")
		}

		if archive {
			callFile.WriteString("Archive: yes\n")
		} else {
			callFile.WriteString("Archive: no\n")
		}

		callFile.Close()

		if schedule != "" {
			now := time.Now()
			w := when.New(nil)
			w.Add(en.All...)
			w.Add(common.All...)
			parsedTime, _ := w.Parse(schedule, now)
			if parsedTime == nil {
				parsedTime = new(when.Result)
				parsedTime.Time = now
			}

			os.Chtimes(spoolFileName, parsedTime.Time, parsedTime.Time)
		}

		// Add call to the outgoing call queue.
		err = os.Rename(spoolFileName, outgoingFileName)
		if err != nil {
			fmt.Println(err)
			s.APISendGeneralResp(w, APIERR, "Unable to move call file into outgoing directory")
			return
		}

		// Send final response.
		s.APISendGeneralResp(w, APIOK, "")
	})
}

func HTTPServe() {
	// Used to reset the app quit timeout for systemd sockets.
	var timeoutReset chan struct{}

	// Create the server.
	httpServer := new(HTTPServer)
	app.httpServer = httpServer

	// Setup the handlers.
	r := http.NewServeMux()
	httpServer.registerHandlers(r)

	// The http server handler will be the mux router by default.
	var handler http.Handler
	handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if app.config.HTTPSystemDSocket {
			timeoutReset <- struct{}{}
		}
		if app.config.HTTPDebug {
			log.Println(req.Method + " " + req.URL.String())
		}
		r.ServeHTTP(w, req)
	})

	// Determine if we're using a systemd socket activation or just a standard listen.
	if app.config.HTTPSystemDSocket {
		done := make(chan struct{})
		quit := make(chan os.Signal, 1)
		timeoutReset = make(chan struct{})

		// On signal, gracefully shut down the server and wait 5
		// seconds for current connection to stop.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Pull existing listener from systemd.
		listeners, err := activation.Listeners()
		if err != nil {
			log.Panicf("Cannot retrieve listeners: %v", err)
		}

		// If we already have a asterisk-outgoing-call-api running, then we shouldn't start...
		if len(listeners) != 1 {
			log.Panicf("Unexpected number of socket activation (%d != 1)", len(listeners))
		}

		server := &http.Server{
			Handler: handler,
		}

		// Upon signal, close out existing connection and quit.
		go func() {
			<-quit
			log.Println("Server is shutting down")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.SetKeepAlivesEnabled(false)
			if err := server.Shutdown(ctx); err != nil {
				log.Panicf("Cannot gracefully shut down the server: %v", err)
			}
			close(done)
		}()

		// 30 minute time out if no connection is received.
		go func() {
			for {
				select {
				case <-timeoutReset:
				case <-time.After(30 * time.Minute):
					close(quit)
				}
			}
		}()

		// Listen on existing systemd socket.
		server.Serve(listeners[0])

		// Wait for existing connections befor exiting.
		<-done
	} else {
		// Get the configuration.
		httpBind := app.config.HTTPBind
		httpPort := app.config.HTTPPort
		if app.flags.HTTPBind != "" {
			httpBind = app.flags.HTTPBind
		}
		if app.flags.HTTPPort != 0 {
			httpPort = app.flags.HTTPPort
		}

		// Start the server.
		log.Println("Starting the http server on port", httpPort)
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", httpBind, httpPort), handler)
		if err != nil {
			log.Fatal(err)
		}
	}
}
