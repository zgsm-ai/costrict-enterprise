/**
 * @file mock utilities
 */

export const successResponse = <T>(data: T) => ({
    code: 0,
    message: 'success',
    data,
});

export const errorResponse = (message: string, code = 1) => ({
    code,
    message,
    data: null,
});
