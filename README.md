# devto-mongodb-hackathon-api
API for my DevTo MongoDB hackathon project.

# Build
Firstly, add you `application_default_credentials.json` to the source of the project. This will enable Google Cloud API.

Then, edit .env file
```bash
mv .env.example .env
```

Edit contents of the .env file.

Finally, you can run the application as you would any other Golang application:
```bash
go build -o devto-mongodb-hackathon github.com/tavsec/devto-mongodb-hackathon-api
```

You can also use Docker:
```bash
docker-compose up -d
```


