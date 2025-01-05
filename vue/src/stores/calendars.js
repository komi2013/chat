import { defineStore } from 'pinia';

export const useCalendarsStore = defineStore({
  id: 'calendars',
  state: () => ({
    calendars: [],
  }),
  actions: {
    insert(data) {
      this.calendars.push(data);
    },
    unshift(data, position = 0) {
      this.calendars.splice(position, 0, data);
    },
    upsert(data) {
      const index = this.calendars.findIndex(calendar => calendar.calendarID === data.calendarID);
      if (index !== -1) {
        this.calendars[index] = data;
      } else {
        this.calendars.push(data);
      }
    },
    update(data, calendarID) {
      const index = this.calendars.findIndex(calendar => calendar.calendarID === calendarID);
      if (index !== -1) {
        this.calendars[index] = data;
      }
    }
  },
});
