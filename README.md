# MSF Pal Alliance Report Generator
This script will query MSF Pal data for an alliance, generate reports, then write them to a [Google Sheet](https://docs.google.com/spreadsheets/d/1p-nLLUjPpkNGiMunDCKryYgRtff8h_P_ag8nK7w7FkI/edit#gid=0).

## Reports
- [War Offense Team Power Average](https://docs.google.com/spreadsheets/d/1p-nLLUjPpkNGiMunDCKryYgRtff8h_P_ag8nK7w7FkI/edit#gid=1338346056)
- [War Defense Team Power Average](https://docs.google.com/spreadsheets/d/1p-nLLUjPpkNGiMunDCKryYgRtff8h_P_ag8nK7w7FkI/edit#gid=1716649418)
- [War Flex Team Power Average](https://docs.google.com/spreadsheets/d/1p-nLLUjPpkNGiMunDCKryYgRtff8h_P_ag8nK7w7FkI/edit#gid=1707425175)

## Usage
Build the Docker image.
```bash
docker build -t msf-pal-alliance-report-generator:latest .
```
Start the container.
```bash
docker run -d msf-pal-alliance-report-generator:latest
```
That's it! The Docker container will now run silently in the background and trigger the script every hour via a crontab.
