import asyncio
import firebase_admin
from firebase_admin import credentials
from firebase_admin import firestore
from azure.eventhub import EventData
from azure.eventhub.aio import EventHubProducerClient
import json

# NOTE: 接続文字列を入れる
EVENT_HUB_CONNECTION_STR = "hoge"
EVENT_HUB_NAME = "micce-search-engine-pre"


async def run():
    cred = credentials.Certificate("/Users/jinzhengpengye/work_space/micce-search-engine/reader/micce-travel-firebase-adminsdk-3fcnp-1e2623f81b.json")
    firebase_admin.initialize_app(cred)
    producer = EventHubProducerClient.from_connection_string(
            conn_str=EVENT_HUB_CONNECTION_STR, eventhub_name=EVENT_HUB_NAME
        )

    db = firestore.client()

    doc_stream = (
        db.collection("Spot").stream()
    )

    batch = await producer.create_batch()
    count = 0
    for doc in doc_stream:
        if count >= 100:
            async with producer:
                print("send data cuz count exceed 100")
                await producer.send_batch(batch)
            batch = await producer.create_batch()
            count = 0
            continue
        elif doc.exists and any(doc.to_dict()):
            data = {"type": "index", "spot_id": doc.to_dict()["id"]}
            json_data = json.dumps(data, indent=2)
            batch.add(EventData(json_data))
            count += 1
        else:
            async with producer:
                print("send data last")
                await producer.send_batch(batch)
    


asyncio.run(run())