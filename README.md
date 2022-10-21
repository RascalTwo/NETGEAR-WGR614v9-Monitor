# NETGEAR WGR614v9 Network Monitor

The NETGEAR WGR614v9 is an old router, therefore while it has network traffic information, it's in the form of network statistics:

![plain network statistics](https://user-images.githubusercontent.com/9403665/128137734-beee8bdd-9c2f-4b34-8b94-42ff12e6fde2.png)

This is both not visually appealing, and not understandable to normal people.

***

That is what this program solves!

Give it the connection information, and you'll be presented with the same data as this network statistics table, but in human-readable formats and with a graph!

https://user-images.githubusercontent.com/9403665/128138759-af37011b-1511-4365-b792-314fb8d62403.mp4

## How It's Made

**Tech Used:** Go, HTML, CSS, JavaScript

First using Go and HTTP Basic Authentication, I logged in and got the network statistics from the router display table, which being called from Go server passed such data to the client-side JavaScript.

Then using JavaScript, I created a graph using the data from the raw data, additionally mirroring the table in a more human-readable format.

## Optimizations

The graphing can be done via a library, offering more customization and better animations, in addition the updating of data can be done via Server-Sent Events or WebSockets, allowing for a more dynamic and responsive experience.

## Lessons Learned

I learned the inner workings of some routers, how easy it was to authenticate and get data from them, and how to use Go to create a server and serve a web page.
