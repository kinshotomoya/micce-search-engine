package micce;

import com.google.api.core.ApiFuture;
import com.google.auth.oauth2.GoogleCredentials;
import com.google.cloud.firestore.Firestore;
import com.google.cloud.firestore.QueryDocumentSnapshot;
import com.google.cloud.firestore.QuerySnapshot;
import com.google.firebase.FirebaseApp;
import com.google.firebase.FirebaseOptions;
import com.google.firebase.cloud.FirestoreClient;

import java.io.IOException;
import java.util.List;
import java.util.concurrent.ExecutionException;

public class Indexer {

    // firestoreにあるSpotのIDは0~9, A~Z, a~zの60通りあるので
    // 60分割でバッチ処理をする
    // firestoreから60分割のデータを取得
    // vespaに格納する
    // その次の60分割のデータをfirestoreから取得
    public static void main(String[] args) throws IOException, ExecutionException, InterruptedException {
        // fire store read
        // 参考：https://github.com/googleapis/java-firestore/blob/main/samples/snippets/src/main/java/com/example/firestore/Quickstart.java
        String projectId = "micce-travel";
        GoogleCredentials credentials = GoogleCredentials.getApplicationDefault();
        FirebaseOptions options = FirebaseOptions.builder()
                .setCredentials(credentials)
                .setProjectId(projectId)
                .build();
        FirebaseApp.initializeApp(options);

        Firestore db = FirestoreClient.getFirestore();
        // TODO: // firestoreにあるSpotのIDは0~9, A~Z, a~zの60通りあるので
        //    // 60分割でバッチ処理をする
        // ↓でprefix matchできる
        ApiFuture<QuerySnapshot> query = db.collection("Spot").orderBy("id").startAt("0").endAt("0" + "\uf8ff").get();

        QuerySnapshot querySnapshot = query.get();
        List<QueryDocumentSnapshot> documents = querySnapshot.getDocuments();
        for (QueryDocumentSnapshot document : documents) {
            System.out.println(document.get("id"));
        }

        // vespa feed
        // 参考: https://github.com/vespa-engine/vespa/blob/master/vespa-feed-client-api/src/test/java/ai/vespa/feed/client/examples/SimpleExample.java
        System.out.println("sss");
    }
}