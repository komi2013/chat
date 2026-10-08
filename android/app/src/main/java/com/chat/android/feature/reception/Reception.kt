package com.chat.android.feature.reception

data class Reception(
    val receptionID: String = "",
    val channelID: String = "",
    val receptionTitle: String = "",
    val adminNames: List<String> = emptyList(),
    val joinNames: List<String> = emptyList(),
    val passcodes: List<Passcode> = emptyList(),
    val asks: List<Ask> = emptyList(),
    val askChoices: List<Ask> = emptyList(),
    val askMultiChoices: List<Ask> = emptyList(),
    val facilities: List<Facility> = emptyList(),
    val openTimes: List<OpenTime> = emptyList(),
    val shifts: List<Shift> = emptyList(),
    val menus: List<Menu> = emptyList(),
    val skills: List<String> = emptyList(),
    val staffSkills: List<StaffSkill> = emptyList(),
    val workStaffNeed: Boolean = false,
    val itemDetails: List<ItemDetail> = emptyList(),
    val waitConfigs: List<WaitConfig> = emptyList(),
    val books: List<Book> = emptyList(),
    val workStaffs: List<WorkStaff> = emptyList(),
    val queues: List<Queue> = emptyList()
)

data class Queue(
    val queueName: String = "",
    val waitingGuest: Int = 0,
    val queuedAt: String = ""
)

data class Book(
    val bookID: String = "",
    val receptionID: String = "",
    val menuID: Int = 0,
    val aliasName: String = "",
    val bookStart: String = "",
    val bookEnd: String = "",
    val answers: List<String> = emptyList()
)

data class WorkStaff(
    val aliasName: String = "",
    val workStart: String = "",
    val workEnd: String = ""
)

data class Passcode(
    val passkey: String = "",
    val usageLimit: Int = 1,
    val passStart: String = "",
    val passEnd: String = ""
)

data class Ask(
    val question: String = "",
    val choices: List<String> = emptyList(),
    val sequence: Int = 0
)

data class Facility(
    val facilityName: String = "",
    val capacity: Int = 0,
    val bookable: Boolean = false
)

data class OpenTime(
    val limitStart: String = "",
    val limitEnd: String = ""
)

data class Shift(
    val aliasNames: List<String> = emptyList(),
    val shiftStart: String = "",
    val shiftEnd: String = "",
    val open: Int = 1,
    val skill: String = "",
    val fix: Boolean = false
)

data class Menu(
    val menuID: Int = 0,
    val menuName: String = "",
    val price: Int = 0,
    val prepaidPrice: Int = 0,
    val needSkill: String = "",
    val needFacility: String = "",
    val specifyNameFlag: Boolean = false,
    val spendMinute: Int = 0,
    val items: List<Int> = emptyList(),
    val paidOptions: List<PaidOption> = emptyList(),
    val freeOptions: List<List<Int>> = emptyList(),
    val freeMultiOptions: List<Int> = emptyList(),
    val forBookType: String = ""
)

data class PaidOption(
    val itemID: Int = 0,
    val price: Int = 0
)

data class StaffSkill(
    val aliasName: String = "",
    val skills: List<String> = emptyList()
)

data class ItemDetail(
    val itemID: Int = 0,
    val itemName: String = "",
    val imgPath: String = "",
    val choices: List<List<String>> = emptyList()
)

data class WaitConfig(
    val guestRange: List<Int> = listOf(1, 5),
    val waitRatio: Int = 0
)

data class ReceptionOrder(
    val receptionOrderID: String = "",
    val seatName: String = "",
    val menuID: Int = 0,
    val menuName: String = "",
    val itemChoices: Map<String, String> = emptyMap(),
    val freeOptions: List<Int> = emptyList(),
    val paidOptions: List<ReceptionOrderPaidOption> = emptyList(),
    val freeMultiOptions: List<Int> = emptyList(),
    val price: Int = 0,
    val items: List<ReceptionOrderItem> = emptyList()
)

data class ReceptionOrderPaidOption(
    val itemID: Int = 0,
    val price: Int = 0
)

data class ReceptionOrderItem(
    val itemID: Int = 0,
    val itemName: String = "",
    val imgPath: String = "",
    val choices: List<List<String>> = emptyList()
)
