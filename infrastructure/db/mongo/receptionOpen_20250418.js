{
    _id: '3eHg',
    admin_names: [
        'mik2'
    ],
    join_names: [
        'ivan1',
        'mik2'
    ],
    passcodes: [
        {
            passkey: 'abc123',
            usage_limit: 10,
            pass_start: '2025-04-15T09:00',
            pass_end: '2025-04-27T18:00'
        },
        {
            passkey: 'guest456',
            usage_limit: 5,
            pass_start: '2025-04-16T08:00',
            pass_end: '2025-04-26T20:00'
        }
    ],
    reception_title: 'サロンの公開用予約リンク',
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
            facility_count: 4,
            facility_name: 'perm'
        }
    ],
    shifts: [
        {
            alias_names: [
                'mik2',
                'ivan1'
            ],
            shift_start: '2025-04-15T10:00',
            shift_end: '2025-04-15T23:00',
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
            shift_start: '2025-04-16T15:00',
            shift_end: '2025-04-16T20:00',
            open: 2,
            role: 'stylist',
            fix: true
        },
        {
            alias_names: [
                'ivan1',
                'mik2'
            ],
            shift_start: '2025-04-17T15:00',
            shift_end: '2025-04-17T20:00',
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
    menus: [
        {
            menu_id: 1,
            menu_name: 'パーマ',
            price: 10000,
            need_skill: 'perm',
            need_facility: 'perm',
            spend_minute: 60
        },
        {
            menu_id: 2,
            menu_name: 'カット',
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
            book_start: '2025-04-16T10:30',
            book_end: '2025-04-16T18:00',
            menu_id: 2
        },
        {
            book_start: '2025-04-18T10:30',
            book_end: '2025-04-18T18:00',
            menu_id: 1,
            people: 5
        },
        {
            book_start: '2025-03-25T16:30',
            book_end: '2025-03-25T17:30',
            menu_id: 1
        },
        {
            book_start: '2025-03-25T18:20',
            book_end: '2025-03-25T19:00',
            answers: [
                'aaa',
                '女',
                '16 ~ 18'
            ],
            menu_id: 1
        },
        {
            book_start: '2025-03-25T15:00',
            book_end: '2025-03-25T16:00',
            answers: [
                'おおお',
                '女',
                '16 ~ 18'
            ],
            menu_id: 1,
            created_at: ISODate('2025-03-22T01:35:02.961Z')
        }
    ],
    queues: [
        {
            waiting_guest: 1,
            queue_name: '小林さん',
            queued_at: '2025-04-014T20:00'
        },
        {
            waiting_guest: 2,
            queue_name: '番号の方がいいかも',
            queued_at: '2025-04-014T20:10'
        }
    ],
    work_staffs: [
        {
            alias_name: 'mik3',
            work_start: '2025-04-05T10:00',
            work_end: '2025-04-05T23:00',
            seq: 2
        },
        {
            alias_name: 'ivan1',
            work_start: '2025-04-05T15:00',
            work_end: '2025-04-05T20:00',
            seq: 2
        },
        {
            alias_name: 'mik2',
            work_start: '2025-04-05T15:00',
            work_end: '2025-04-05T20:00',
            seq: 3
        }
    ],
    wait_configs: [
        {
            guest_range: [
                1,
                1
            ],
            wait_ratio: 3
        },
        {
            guest_range: [
                2,
                3
            ],
            wait_ratio: 10
        },
        {
            guest_range: [
                4,
                7
            ],
            wait_ratio: 20
        },
        {
            guest_range: [
                8,
                0
            ],
            wait_ratio: 40
        }
    ]
}