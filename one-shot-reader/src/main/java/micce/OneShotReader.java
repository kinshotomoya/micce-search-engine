package micce;

import com.google.api.core.ApiFuture;
import com.google.cloud.firestore.Firestore;
import com.google.cloud.firestore.QueryDocumentSnapshot;
import com.google.cloud.firestore.QuerySnapshot;

import java.util.List;
import java.util.concurrent.ExecutionException;
import java.util.stream.Collectors;

/*
全件更新の際に単発でpubsubにpublishするバッチ
 */
public class OneShotReader {
    public static void main(String[] args) {
        // firestoreにあるSpotのIDは0~9, A~Z, a~zの60通りあるので
        // 60分割でバッチ処理をする
        // firestoreから60分割のデータを取得
        // vespaに格納する
        // その次の60分割のデータをfirestoreから取得
        final String[] spotIdPrefix = {
                "0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
                "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
                "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"
        };


        // NOTE: FireStoreClientはCloseable interfaceを実装しているので、try-with-resourcesを利用できる
        try(
                FireStoreClient fireStore = new FireStoreClient("micce-travel");
        ) {
            for(String id : spotIdPrefix) {
                List<FireStoreSpot> spotList = getSpotFromFireStore(fireStore.db, id);
                publishToQueue(spotList);
            }
        } catch (Exception e) {
            System.out.println(e.getMessage());
        }


    }

    // TODO: 
    // pubsubにpublishする
    static void publishToQueue(List<FireStoreSpot> spotList) {
        // pubsub client利用
        // FireStoreSpotをjson形式にする必要がありそう
    }


    // fire store read
    // 参考：https://github.com/googleapis/java-firestore/blob/main/samples/snippets/src/main/java/com/example/firestore/Quickstart.java
    static List<FireStoreSpot> getSpotFromFireStore(Firestore db, String id) throws ExecutionException, InterruptedException {
        ApiFuture<QuerySnapshot> query = db.collection("Spot").orderBy("id").startAt(id).endAt(id + "\uf8ff").get();
        QuerySnapshot querySnapshot = query.get();
        List<QueryDocumentSnapshot> documents = querySnapshot.getDocuments();
        return documents.stream().map(doc -> new FireStoreSpot(
                doc.getString("id"),
                doc.getString("name"),
                doc.getString("koreaName"),
                doc.getDouble("latitude"),
                doc.getDouble("longitude")
                )).collect(Collectors.toList());
    }


}