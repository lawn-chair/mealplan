import { parse, format } from 'date-fns';

export function formatDate(date: string) {
    return format(parse(date, "yyyy-MM-dd", new Date), "MMMM d, yyyy");
}

export function formatDateLong(date: string) {
    return format(parse(date, "yyyy-MM-dd", new Date), "EEEE, MMMM d, yyyy");
}  
