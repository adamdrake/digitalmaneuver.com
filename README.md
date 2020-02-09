# Overview

This is the frontend and backend code that powers [Digital Maneuver](https://digitalmaneuver.com).

# Frontend

The frontend is just a few HTML and CSS files.  It's designed to load very quickly on mobile devices with limited bandwidth.  CSS is modeled after [BMFW](https://bestmotherfucking.website).

# Backend

The backend is a hackish service that can facilitate adding emails to a SendGrid contact list.  There are no tests but it seems to work, so YMMV.  This was not designed for production, low-latency, etc.

The backend can run on a local machine, VPS, etc. or also as an AWS Lambda function.  If you set the `RUN_AS_LAMBDA` environment variable to be `TRUE` then the code path will behave as though it's running on AWS Lambda.  Rate limiting is handled via AWS API Gateway.
