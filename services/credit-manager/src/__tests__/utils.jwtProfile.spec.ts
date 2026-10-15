import { describe, expect, it } from 'vitest';
import { readJwtProfile } from '@/utils/jwtProfile';

function token(claims: unknown): string {
    const binary = Array.from(new TextEncoder().encode(JSON.stringify(claims)), (byte) =>
        String.fromCharCode(byte),
    ).join('');
    const payload = btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
    return `header.${payload}.signature`;
}

const empty = { userName: '', userId: '', universalId: '', employeeNumber: '' };

describe('JWT profile display', () => {
    it('decodes UTF-8 names and unpadded base64url fields', () => {
        expect(
            readJwtProfile(
                token({
                    display_name: '张三 😀',
                    sub: 'user-123',
                    universal_id: 'quota-456',
                    employee_number: '007',
                }),
            ),
        ).toEqual({
            userName: '张三 😀',
            userId: 'user-123',
            universalId: 'quota-456',
            employeeNumber: '007',
        });
    });

    it('allows missing employee number and ignores unsupported field types', () => {
        expect(
            readJwtProfile(
                token({ display_name: 'Alice', sub: 'user-1', universal_id: 'quota-1' }),
            ),
        ).toEqual({
            userName: 'Alice',
            userId: 'user-1',
            universalId: 'quota-1',
            employeeNumber: '',
        });
        expect(
            readJwtProfile(
                token({ display_name: {}, sub: [], universal_id: null, employee_number: false }),
            ),
        ).toEqual(empty);
    });

    it('falls back for missing, malformed or non-object payloads', () => {
        for (const value of [
            undefined,
            '',
            'opaque-token',
            'header.@@.signature',
            'header.bm90LWpzb24.signature',
            token(null),
            token([]),
        ]) {
            expect(readJwtProfile(value)).toEqual(empty);
        }
    });
});
