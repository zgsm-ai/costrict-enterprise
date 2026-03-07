/**
 * @file util date
 */
import dayjs from 'dayjs';

export const formatDate = (
    dateInput: string,
    formatStr = 'YYYY-MM-DD HH:mm:ss',
    invalidDateFallback = '',
) => {
    // If the input is null, undefined, or an empty string, return the fallback value directly
    if (dateInput === null || dateInput === undefined || dateInput === '') {
        return invalidDateFallback;
    }

    // Create a Day.js object
    const date = dayjs(dateInput);

    // Check if the Day.js object is valid
    // If dateInput is an unparsable string, dayjs(dateInput) will still return a Day.js object,
    // but its isValid() method will return false
    if (!date.isValid()) {
        console.warn(
            `Invalid date input: ${dateInput}, returning fallback value: ${invalidDateFallback}`,
        );
        return invalidDateFallback;
    }

    // Format and return the date
    return date.format(formatStr);
};
