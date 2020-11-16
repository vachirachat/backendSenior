# backendSenior

## Setup
1. Install go 
2. get all go-denpendency ## go get -d -v
3. run ## go run main.go 
4. run MongoDB image in local ## docker run -d -p 27017:27017 mongo


// Test socket API 

# API For test 
Follow step

	# 1--> Post http://localhost:8080/dev/v1/createroom
	{
		"RoomName": "TestSocket"   ,       
		"RoomType": "PRIVATE"       ,  
		"ListUser": []
	}
	* GET http://localhost:8080/dev/v1/room
	look like ::
		{
		    "roomId": "5fb261d8584cb606941cab70",
		    "roomName": "TestSocket",
		    "roomType": "PRIVATE",
		    "listUser": []
		}

	# 2--> step add new 3 users
	    2.1 -> Post http://localhost:8080/api/v1/signup
	    {
		"RoomName": "TestSocket"   ,       
		"RoomType": "PRIVATE"       ,  
		"ListUser": []
	    }

    2.2 -> Post http://localhost:8080/api/v1/signup
    {

    }

    2.3 -> Post http://localhost:8080/api/v1/signup
    {
        "RoomName": "TestSocket"   ,       
        "RoomType": "PRIVATE"       ,  
        "ListUser": []
    }

    2.4 -> * See result with GET http://localhost:8080/dev/v1/user
    should look like 
        {
            {
                "userID": "5fb263212a02b5d2bf3b4d8a",
                "name": "test-socket-1",
                "email": "test-socket-1@test.com",
                "password": "$2a$10$AwC7ky2EzUZHTCN3NOHzeO8D3vLbrGGEXEdPxuVEKsdH7ITMlXV8S",
                "room": [],
                "userType": "USER"
            },
            {
                "userID": "5fb2636c2a02b5d2bf3b4da0",
                "name": "test-socket-2",
                "email": "test-socket-2@test.com",
                "password": "$2a$10$14c/PzNOJQdE5OVRDG5Sq.FNOU8EnJW2LS.oLD/2D038ENPVPiynO",
                "room": [],
                "userType": "USER"
            },
            // may use to check join room
            {
                "userID": "5fb263702a02b5d2bf3b4da7",
                "name": "test-socket-3",
                "email": "test-socket-3@test.com",
                "password": "$2a$10$DYK9oCOn4CqQjoToJkPuruupqllKM.GxVsHk5kp52wkxvhTPTTcP2",
                "room": [],
                "userType": "USER"
            }

	# 3--> POST http://localhost:8080/dev/v1/addmembertoroom
	{
	    "roomID" : "5fb261d8584cb606941cab70",
		"ListUser" : ["test-socket-1 ID","test-socket-1 2 ID"]
	}
	
	# 4--> Call GET http://localhost:8080/dev/v1/room && GET http://localhost:8080/dev/v1/user you will setup for start socket
