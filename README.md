# CloudDB
Cloud DB is cloud based data storage extension of GoldenCheetah (available in GoldenCheetah Version 3.4).

It is planned to provide Cloud Database for GoldenCheetah artifacts

1. Charts (including chart specific metrics) (done - available in GC 3.4 onwards)
2. User Metrics (Not planned yet)
3. Layouts (Not planned yet)
4. Workouts (Not planned yet)

... more artifacts to come with new features being added to GoldenCheetah

GoldenCheetah plans to provide build-in functions to post artifacts to the Cloud DB - which
are useful for other users. As well as functions so that every user can view and retrieve
artifacts from the Cloud DB to use in their own local GoldenCheetah installation.

****

What is Cloud DB not:

- it's not a cloud storage for User Data like activities,...
- it's not a data sharing platform for athlete's and their trainers


****

Technical details - CloudDB is a Google App Engine application executed on Google
Infrastructure. The APP can only be accessed through the GoldenCheetah application which
therefore must be installed locally on your PC. There is not Web based access to any
of the CloudDB functions provided or even planned.

Depending on the overall access volume to the application there may be limitations
on the availability of this service - e.g. regarding the number of artifact up-/downloads,
or even on 7x24 availability.

***

# Disclaimer

All concepts/ideas/plans can be changed or even stopped at any point of time without
notice to anyone (user, developer,...). There is not commitment that the CloudDB service
will be available and operated all time.

The decision to run the service is subject to the GoldenCheetah development team
and the service can be stopped or limited at any time - without prior notice.

