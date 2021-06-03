# Process Ping Source Events

In this example we will send events from ping source to a broker backed by KafkaTopicChannel and will create triggers to send the ping source events to a knative service called 'cev'

### PreReq

Please have a kafka topic handy. Once you have the topic, (if needed) create the auth secrets of kafka broker like below

```
$ kubectl create secret generic cacert --from-file=caroot.pem
secret/cacert created

$ kubectl create secret tls kafka-secret --cert=certificate.pem --key=key.pem
secret/key created
```

### Logging and Tracing 
Create the logging and tracing config map in the namespace you want the channel to be created. This is a one time activity per namespace

```
kubectl apply -f obs.yaml
```

### Channel CM

```
kubectl apply -f channel-cm.yaml
```

### Knative service

Create the knative service to recieve events

```
kubectl apply -f ksvc.yaml
```

Open the cloud events viewer in the browser by visting the url `http://ksvc-url/index.html`

### Broker and Trigger

```
kubectl apply -f broker-trigger.yaml
```

### Ping Source

Time to send events

```
kubectl apply -f pingsource.yaml
```

Now every 1 minute the cloud events viewer should display the event
