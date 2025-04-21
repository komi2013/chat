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
            alias_names: [
                'mik2',
                'ivan1'
            ],
            shift_start: '2025-03-15T10:00',
            shift_end: '2025-03-15T23:00',
            open: 1,
            role: 'stylist',
            fix: true
        },
        {
            alias_names: [
                'ivan1',
                'mik2',
                'mik3'
            ],
            shift_start: '2025-03-16T15:00',
            shift_end: '2025-03-16T20:00',
            open: 2,
            role: 'stylist',
            fix: true
        },
        {
            alias_names: [
                'ivan1',
                'mik2'
            ],
            shift_start: '2025-03-17T15:00',
            shift_end: '2025-03-17T20:00',
            open: 1,
            role: 'helper'
        }
    ],
    open_times: [
        {
            limit_start: '2025-03-19T10:00',
            limit_end: '2025-03-19T20:00'
        },
        {
            limit_start: '2025-03-14T15:00',
            limit_end: '2025-03-14T20:00'
        }
    ],
    services: [
        {
            service_id: 1,
            service_name: 'パーマ',
            price: 10000,
            need_skill: 'perm',
            need_facility: 'perm',
            spend_minute: 60
        },
        {
            service_id: 2,
            service_name: 'カット',
            need_skill: 'cut',
            price: 3000,
            spend_minute: 60
        }
    ],
    skills: [
        'cut',
        'perm'
    ],
    staff_skills: [
        {
            alias_name: 'ivan1',
            skills: [
                'perm',
                'cut'
            ]
        },
        {
            alias_name: 'mik2',
            skills: [
                'perm'
            ]
        }
    ],
    work_staff_need: true,
    books: [
        {
            book_start: '2025-03-22T10:30',
            book_end: '2025-03-22T11:00',
            service_id: 2
        },
        {
            book_start: '2025-03-22T16:00',
            book_end: '2025-03-22T17:00',
            service_id: 1
        },
        {
            book_start: '2025-03-22T16:30',
            book_end: '2025-03-22T17:30',
            service_id: 1
        }
    ],
    work_staffs: [
        {
            alias_name: 'mik3',
            work_start: '2025-03-22T10:00',
            work_end: '2025-03-22T23:00',
            seq: 2
        },
        {
            alias_name: 'ivan1',
            work_start: '2025-03-22T15:00',
            work_end: '2025-03-22T20:00',
            seq: 2
        },
        {
            alias_name: 'mik2',
            work_start: '2025-03-22T15:00',
            work_end: '2025-03-22T20:00',
            seq: 3
        }
    ]
}