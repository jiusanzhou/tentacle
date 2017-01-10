# Tentacles

> A new mode of distributed proxy.

## What we want?

We want plenty of IPs for web-crawlling or something else.

## How we get?

The only way that can make that is use ADSL net.
Yes, we use different cities' home's ADSL to serve for our fetcher.
Change the IP at anytime if we want.

## Why is this?

Home ADSL net can not be used as a server.
If we make it as a normal proxy server,
thing we need to do not just setup the proxy server,
alse contains tunning the net to make this service can be attached
by user on the internet(always is our program).

We have a way to tunning out of a net, like Ngrok.
But the thing is we want higher performance,
hance we need to change a way to finished our target.

The less processing the higher performance.

**Easy way**

```
[Request]---->[Tunnel]---->[Tunnel]----[Proxy]---->[Target]
```

**Tentacles' way**

```
[Request]---->[Tentacles]---->[Target]
```

## Actually Tentacles is?

Tentacles is proxy server[socket5, http] with different exports.
The export can also be used as a client, every single tentacle is for all, all tentacles are for every single one.