#!/bin/bash

# Start apache. It gets grumpy about PID files pre-existing
rm -f /usr/local/apache2/logs/httpd.pid
httpd -DFOREGROUND