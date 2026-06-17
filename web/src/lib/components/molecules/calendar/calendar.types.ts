export const calendarModes = ['single', 'range'] as const;

export type CalendarMode = (typeof calendarModes)[number];
