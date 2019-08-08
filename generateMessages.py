#!/bin/python3
import requests
import random
import uuid

url = "http://localhost:8000/hcfse/post"

politePayload = "\"username\": \"Polite Guy {}\",\"content\": \"{} {}\""
rudePayload = "\"username\": \"Rude Guy {}\",\"content\": \"I wanna bad your liar\""

headers = {
    'Content-Type': "application/json",
    'cache-control': "no-cache"
    }

for _ in range(100):
    response = requests.request("POST", url, data="{"+rudePayload.format(random.randint(1, 50))+"}", headers=headers)
    print(response.text)
    for _ in range(4):
        response = requests.request("POST", url, data="{"+politePayload.format(random.randint(1, 1000), uuid.uuid1(), uuid.uuid1())+"}", headers=headers)
        print(response.text)

# Uncomment for example of someone getting banned
# for _ in range(11):
#     response = requests.request("POST", url, data="{"+rudePayload.format("Who Is Gonna Get Banned")+"}", headers=headers)
#     print(response.text)