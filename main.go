/*
Copyright 2014 Rohith All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"net/http"
	"encoding/json"
	"net"
	"strings"
	"errors"

	"github.com/jbdalido/gomarathon"
	"github.com/golang/glog"
	"fmt"
)

var options struct {
	/* the endpoint for marathon */
	marathon_url string
	/* the interface we should bind to */
	event_interface string
	/* the port we should listen on */
	event_port int

	}

func init() {
	flag.StringVar(&options.marathon_url,"marathon", "http://localhost:8080", "the endpoint url for marathon")
	flag.StringVar(&options.event_interface,"interface", "eth0", "the interface we should bind to")
	flag.IntVar(&options.event_port,"port", 8080, "the port we should use for listening for events")
}

func main() {
	flag.Parse()

	if options.marathon_url == "" {
		glog.Errorf("You have not specified the marathon url to connect, exitting")
		return
	}

	/* step: register for the events */
	glog.Infof("Registering for marathon events on endpoint: %s", options.marathon_url)
	if client, err := gomarathon.NewClient(options.marathon_url, nil); err != nil {
		glog.Errorf("Failed to connect to the marathon endpoint: %s, error: %s", options.marathon_url, err )
		return
	} else {
		/* step: we need to get our ip address */
		ip_address, err := GetInterfaceAddress(options.event_interface)
		if err != nil {
			glog.Errorf("Failed to get the ip address from interface: %s, error: %s", options.event_interface, err )
			return
		}
		if response, err := client.GetEventSubscriptions(); err != nil {
			glog.Errorf("Failed to get a list of current subscriptions, error: %s", err )
			return
		} else {
			glog.Infof("response: %v", response)
		}

		registered_url := fmt.Sprintf("http://%s:%d/callback", ip_address, options.event_port)
		glog.Infof("Registering local events endpoint: %s", registered_url)
		if _, err := client.RegisterCallbackURL(registered_url); err != nil {
			glog.Errorf("Failed to register for events from marathon: %s, error: %s", options.marathon_url, err )
			return
		}

		/* step: listen for events */
		http.HandleFunc("/", HandleMarathonEvent )
		http.ListenAndServe(fmt.Sprintf("%s:8080", ip_address, options.event_port), nil)

	}
}


func GetInterfaceAddress(interface_name string) (string, error) {
	glog.Infof("Attempting to grab the ipaddress of interface: %s", interface_name)
	if interfaces, err := net.Interfaces(); err != nil {
		glog.Errorf("Unable to get the proxy ip address, error: %s", err)
		return "", err
	} else {
		for _, iface := range interfaces {
			/* step: get only the interface we're interested in */
			if iface.Name == interface_name {
				glog.V(6).Infof("Found interface: %s, grabbing the ip addresses", iface.Name)
				addrs, err := iface.Addrs()
				if err != nil {
					glog.Errorf("Unable to retrieve the ip addresses on interface: %s, error: %s", iface.Name, err)
					return "", err
				}
				/* step: return the first address */
				if len(addrs) > 0 {
					return strings.SplitN(addrs[0].String(), "/", 2)[0], nil
				} else {
					glog.Fatalf("The interface: %s has no ip address", interface_name)
				}
			}
		}
	}
	return "", errors.New("Unable to determine or find the interface")
}

func HandleMarathonEvent(writer http.ResponseWriter, request *http.Request) {
	/* step: create the buffer */
	buffer := make([]byte, request.ContentLength)
	/* step: read in the content */
	if _, err := request.Body.Read(buffer); err != nil {
		glog.Errorf("Failed to reading the content from request, error: %s", err )
		return
	}
	/* step: attempt to unmarshal the data */
	var data map[string]interface {}
	if err := json.Unmarshal(buffer, &data); err != nil {
		glog.Errorf("Failed to unmarshall the json event, error: %s", err )
	} else {
		glog.Infof("Event: %v", data )
	}
}

