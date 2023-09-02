package micce;

import com.google.api.core.ApiFuture;
import com.google.cloud.firestore.Firestore;
import com.google.cloud.firestore.QueryDocumentSnapshot;
import com.google.cloud.firestore.QuerySnapshot;

import java.util.List;
import java.util.concurrent.ExecutionException;
import java.util.stream.Collectors;

/*
変更のあったfirestoreデータのみ取得し、pubsubにpublishするシステム
 */
public class Reader {
    public static void main(String[] args) {
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

    // pubsubにpublishする
    // 参考: // https://firebase.google.com/docs/firestore/query-data/listen?hl=ja
    static void publishToQueue(List<FireStoreSpot> spotList) {

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