go build -o mumail  && \
zip mumail.zip mumail && \
rm mumail && \
aws lambda update-function-code --function-name mumail --zip-file fileb://mumail.zip && \
rm mumail.zip
