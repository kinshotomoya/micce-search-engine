package micce;

public class FireStoreSpot {
    String id;
    String name;
    String koreaName;
    Double lat;
    Double lon;

    public FireStoreSpot(String id, String name, String koreaName, Double lat, Double lon) {
        this.id = id;
        this.name = name;
        this.koreaName = koreaName;
        this.lat = lat;
        this.lon = lon;
    }

}
