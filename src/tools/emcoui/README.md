## Local setup

for running the app in a local setup first install the dependencies by running `npm install`.
Then run `startup.sh`

## Production build

for creating a production build, run `npm run build`. A production ready build will be available at /build directory

## Available scripts

### `startup.sh`

This script basically calls npm start.
This script runs the app in the development mode.<br />
Before running the script update the backend address if backend is not running locally.
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

The page will reload if you make edits.<br />
You will also see any lint errors in the console.

### `npm run build`

Builds the app for production to the `build` folder.<br />
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.<br />

## Building docker image

To build a docker image run the below command
`docker build -t image_name:version .`
