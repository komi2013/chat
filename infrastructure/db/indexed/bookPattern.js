
const bookPattern = {
      "adminGroup": "group1",
      "bookPatternID": "ncsW",
      "joinNames": [
        "ivan1",
        "mik2"
      ],
      "parentID": "6736d3fd24a8745b9c9f4112",
      "skills": [
        "cut",
        "perm"
      ],
      "shiftStaffs": [
        {
          "shiftTitle": "フルタイム、スタイリスト設定",
          "date": "2025-03-09",
          "shiftStart": "10:00",
          "shiftEnd": "15:00",
          "open": 1,
          "role": "stylist",
          "aliasNames": [
            "mik2"
          ]
        },
        {
          "shiftTitle": "フルタイム、スタイリスト設定",
          "date": "2025-03-08",
          "shiftEnd": "20:00",
          "shiftStart": "15:00",
          "open": 1,
          "role": "stylist",
          "aliasNames": [
            "mik2", "ivan1"
          ]
        },
        {
          "shiftTitle": "フルタイム、スタイリスト設定",
          "date": "2025-03-07",
          "shiftEnd": "20:00",
          "shiftStart": "15:00",
          "open": 1,
          "role": "stylist",
          "aliasNames": [
            "ivan1", "mik2"
          ]
        }
      ]
    }


upsertIDB(bookPattern, 'bookPattern', 'bookPatternID', bookPattern.bookPatternID);