# fetch-wf-tasks

Go program to fetch pending wf tasks that your user is owner, and e-mail them to desired contacts

### dependecies
[workfront](https://github.com/renatovassao/workfront)

### configuration
You need to create a file named "config.json" in the same directory as main.go in the following format:

````json
{
        "Workfront": {
            "User": "...",
            "Pass": "..."
        },
        "Mail": {
            "Server": "mail.example.com",
            "Port": 25,
            "From": {
                "User": "...",
                "Pass": "..."
            },
            "To": ["example@example.com", "example2@example2.com"]
        }
}
````
