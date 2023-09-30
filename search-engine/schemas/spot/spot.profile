rank-profile spot inherits default {
    # https://docs.vespa.ai/en/geo-search.html#ranking-from-a-position-match
    first-phase {
        expression: distance(spot_geo_location)
    }
}
