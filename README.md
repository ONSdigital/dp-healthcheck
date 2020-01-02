dp-healthcheck
================

A health check git repository for DP

### Getting started

Read the [Health Check Specification](https://github.com/ONSdigital/dp/blob/master/standards/HEALTH_CHECK_SPECIFICATION.md) for details.

#### How to use
1. Add Health Check library to an app

2. Create an array of Health Check clients by calling `NewClient()` passing in the following:

- An optional RCHTTP clienter; if none is provided one will be created
- A function that implements the `Checker` interface

3. Call `Create()` passing in the following:

- Versioning Information
- Critical time duration; time to wait for dependent apps critical unhealthy status to make current app unhealthy- Time Interval to run health checks on dependencies
- Clients; An array of clients created in the previous step

4. Optionally call `AddClient` on the healthcheck to add additional clients, note this can only be done prior to `Start()` being called

5. Call `Start()` on the healthcheck

### Configuration

Configuration of the health check takes place via arguments passed to the `Create()` function

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2019, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
