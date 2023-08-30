package micce;

import com.google.auth.oauth2.GoogleCredentials;
import com.google.cloud.firestore.Firestore;
import com.google.firebase.FirebaseApp;
import com.google.firebase.FirebaseOptions;
import com.google.firebase.cloud.FirestoreClient;

import java.io.Closeable;
import java.io.IOException;

public class FireStoreClient implements Closeable {
    String projectId;
    Firestore db;


    public FireStoreClient(String projectId) throws IOException {
        this.projectId = projectId;
        GoogleCredentials credentials = GoogleCredentials.getApplicationDefault();
        FirebaseOptions options = FirebaseOptions.builder()
                .setCredentials(credentials)
                .setProjectId(projectId)
                .build();
        FirebaseApp.initializeApp(options);
        this.db = FirestoreClient.getFirestore();
    }

    @Override
    public void close() throws IOException {
        try {
            this.db.close();
        } catch (Exception e) {
            System.out.println("firestore clientのcloseに失敗");
            throw new RuntimeException(e);
        }
    }
}
