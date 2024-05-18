import asyncio
import json
import time
from datetime import datetime, timezone, timedelta

import firebase_admin
from firebase_admin import credentials
from firebase_admin import firestore
from azure.eventhub import EventData
from azure.eventhub.aio import EventHubProducerClient
from google.cloud.firestore_v1.base_query import FieldFilter

# NOTE: 実行前に接続文字列を入れる
# 以下リンク先の「接続文字列 – 主キー」の値
# https://portal.azure.com/#@micceappgmail.onmicrosoft.com/resource/subscriptions/e5c0c1b7-761e-4239-9b49-ef8ea9655de6/resourceGroups/micce_resource_group/providers/Microsoft.EventHub/namespaces/micce-search-engine/saskey
EVENT_HUB_CONNECTION_STR = "hoge"
EVENT_HUB_NAME = "micce-search-engine-pre"


async def run():
    cred = credentials.Certificate("/Users/jinzhengpengye/work_space/micce-search-engine/reader/micce-travel-firebase-adminsdk-3fcnp-1e2623f81b.json")
    firebase_admin.initialize_app(cred)
    producer = EventHubProducerClient.from_connection_string(
            conn_str=EVENT_HUB_CONNECTION_STR, eventhub_name=EVENT_HUB_NAME
        )

    db = firestore.client()

    # NOTE: 全件はfirebaseのrate limit的な規制に引っかかるのでできるだけ日付け指定する
    # 以下は日本時間2024/01/01 0:00:00の例
    dt = datetime(2023, 10, 1, 00, 00, 00, tzinfo=timezone(timedelta(hours=9)))

    doc_stream = (
        db.collection("Spot")
        .where(filter=FieldFilter("createdAt", ">=", dt))
        .stream()
    )

    batch = await producer.create_batch()
    count = 0
    for doc in doc_stream:
        if count >= 1000:
            async with producer:
                print("send data cuz count exceed 1000")
                await producer.send_batch(batch)
            batch = await producer.create_batch()
            count = 0
            time.sleep(10)
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