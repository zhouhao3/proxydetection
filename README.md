# proxydetection
proxydetection is a tool that detects whether the specified proxy can connect to the specified URL, and will write the last unreachable URL in the result-url file.

# User Stories

* glider-log

  After deploying the glider to obtain the relevant glider log file, this tool can directly process the log file and obtain the unreachable URL.
  ```
  ./proxydetection PORXY FILEPATH
  ```

 * url-file
 
   You can directly write the URL to be tested into a file for testing.
   ```
   ./proxydetection --file-type url-file PORXY FILEPATH
   ```
