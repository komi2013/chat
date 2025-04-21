{
    _id: ObjectId('6736d3fd24a8745b9c9f4112'),
    admin_group: 'group1',
    join_names: [
        'ivan1',
        'mik2'
    ],
    book_title: 'サロンの公開用予約リンク',
    asks: [
        '名前は？'
    ],
    ask_choices: [
        [
            '性別は？',
            '男',
            '女',
            'その他'
        ]
    ],
    ask_multi_choices: [
        [
            '何歳ですか？',
            '~ 15',
            '16 ~ 18',
            '19 ~ 25',
            '26 ~ 35',
            '35 ~'
        ]
    ],
    facilities: [
        {
            facility_count: 1,
            facility_name: 'perm'
        }
    ],
    shifts: [
        {
            shift_title: 'フルタイム、スタイリスト設定',
            date: '2025-03-09',
            shift_start: '10:00',
            shift_end: '15:00',
            open: 1,
            role: 'stylist',
            alias_names: [
                'mik2'
            ]
        },
        {
            shift_title: 'フルタイム、スタイリスト設定',
            date: '2025-03-08',
            shift_end: '20:00',
            shift_start: '15:00',
            open: 1,
            role: 'stylist',
            alias_names: [
                'mik2',
                'ivan1'
            ]
        },
        {
            shift_title: 'フルタイム、スタイリスト設定',
            date: '2025-03-07',
            shift_end: '20:00',
            shift_start: '15:00',
            open: 1,
            role: 'stylist',
            alias_names: [
                'ivan1',
                'mik2'
            ]
        }
    ],
    times: [
        {
            date: '2025-03-09',
            limit_start: '10:00',
            limit_end: '20:00',
            books: [
                {
                    book_start: '12:00',
                    book_end: '13:00',
                    answers: [
                        '小松',
                        '35',
                        '2'
                    ],
                    service_id: 2
                },
                {
                    book_start: '13:00',
                    book_end: '14:00',
                    answers: [
                        '小松',
                        '35',
                        '2'
                    ],
                    service_id: 1
                },
                {
                    book_start: '11:00',
                    book_end: '12:00',
                    answers: [],
                    service_id: 1
                }
            ],
            work_staffs: [
                {
                    alias_name: 'mik2',
                    work_start: '09:00',
                    work_end: '18:00',
                    skills: [
                        'cut'
                    ],
                    seq: 1,
                    'delete': false
                }
            ]
        },
        {
            date: '2025-03-08',
            limit_start: '10:00',
            limit_end: '20:00',
            work_staffs: [
                {
                    alias_name: 'mik2',
                    work_start: '09:00',
                    work_end: '18:00',
                    skills: [
                        'cut'
                    ],
                    seq: 1,
                    'delete': false
                }
            ]
        }
    ],
    services: [
        {
            id: 1,
            service_name: 'パーマ',
            price: 10000,
            need_skill: 'perm',
            need_facility: 'perm',
            spend_minute: 60
        },
        {
            id: 2,
            service_name: 'カット',
            need_skill: 'cut',
            price: 3000,
            spend_minute: 60
        }
    ],
    skills: [
        'cut',
        'perm'
    ]
}