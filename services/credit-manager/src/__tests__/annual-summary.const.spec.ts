import { describe, it, expect } from 'vitest';
import {
    BACKEND_TYPE_MAP,
    TYPE_IMAGE_MAP,
    EXPORT_IMAGE_MAP,
    TYPE_TEXT_MAP,
    TYPE_VIDEO_MAP,
    DEFAULT_RESULT_TYPE,
} from '@/views/AnnualSummary/const';

describe('AnnualSummary Constants', () => {
    describe('BACKEND_TYPE_MAP', () => {
        it('should map backend types to internal types correctly', () => {
            expect(BACKEND_TYPE_MAP.pioneer).toBe('first');
            expect(BACKEND_TYPE_MAP.high_freq).toBe('agent');
            expect(BACKEND_TYPE_MAP.problem_solver).toBe('bug');
            expect(BACKEND_TYPE_MAP.efficiency).toBe('speed');
        });

        it('should have 4 backend type mappings', () => {
            expect(Object.keys(BACKEND_TYPE_MAP)).toHaveLength(4);
        });
    });

    describe('TYPE_IMAGE_MAP', () => {
        it('should have image paths for all internal types', () => {
            expect(TYPE_IMAGE_MAP.agent).toBeDefined();
            expect(TYPE_IMAGE_MAP.bug).toBeDefined();
            expect(TYPE_IMAGE_MAP.first).toBeDefined();
            expect(TYPE_IMAGE_MAP.speed).toBeDefined();
        });

        it('should have non-empty string values', () => {
            Object.values(TYPE_IMAGE_MAP).forEach((value) => {
                expect(typeof value).toBe('string');
                expect(value.length).toBeGreaterThan(0);
            });
        });
    });

    describe('EXPORT_IMAGE_MAP', () => {
        it('should have export image paths for all internal types', () => {
            expect(EXPORT_IMAGE_MAP.agent).toBeDefined();
            expect(EXPORT_IMAGE_MAP.bug).toBeDefined();
            expect(EXPORT_IMAGE_MAP.first).toBeDefined();
            expect(EXPORT_IMAGE_MAP.speed).toBeDefined();
        });

        it('should match TYPE_IMAGE_MAP keys', () => {
            expect(Object.keys(EXPORT_IMAGE_MAP)).toEqual(Object.keys(TYPE_IMAGE_MAP));
        });
    });

    describe('TYPE_TEXT_MAP', () => {
        it('should have text descriptions for all internal types', () => {
            expect(TYPE_TEXT_MAP.bug).toBeDefined();
            expect(TYPE_TEXT_MAP.first).toBeDefined();
            expect(TYPE_TEXT_MAP.speed).toBeDefined();
            expect(TYPE_TEXT_MAP.agent).toBeDefined();
        });

        it('should have meaningful descriptions', () => {
            Object.entries(TYPE_TEXT_MAP).forEach(([key, value]) => {
                expect(typeof value).toBe('string');
                expect(value.length).toBeGreaterThan(20);
                expect(key).toBeTruthy();
            });
        });
    });

    describe('TYPE_VIDEO_MAP', () => {
        it('should have video paths for all internal types', () => {
            expect(TYPE_VIDEO_MAP.agent).toBeDefined();
            expect(TYPE_VIDEO_MAP.bug).toBeDefined();
            expect(TYPE_VIDEO_MAP.first).toBeDefined();
            expect(TYPE_VIDEO_MAP.speed).toBeDefined();
        });

        it('should match TYPE_IMAGE_MAP keys', () => {
            expect(Object.keys(TYPE_VIDEO_MAP)).toEqual(Object.keys(TYPE_IMAGE_MAP));
        });
    });

    describe('DEFAULT_RESULT_TYPE', () => {
        it('should be speed', () => {
            expect(DEFAULT_RESULT_TYPE).toBe('speed');
        });

        it('should be a valid type key', () => {
            expect(TYPE_IMAGE_MAP[DEFAULT_RESULT_TYPE]).toBeDefined();
            expect(TYPE_TEXT_MAP[DEFAULT_RESULT_TYPE]).toBeDefined();
            expect(TYPE_VIDEO_MAP[DEFAULT_RESULT_TYPE]).toBeDefined();
        });
    });

    describe('Consistency checks', () => {
        it('should have consistent internal type keys across all maps', () => {
            const internalTypes = ['agent', 'bug', 'first', 'speed'];

            internalTypes.forEach((type) => {
                expect(TYPE_IMAGE_MAP[type]).toBeDefined();
                expect(EXPORT_IMAGE_MAP[type]).toBeDefined();
                expect(TYPE_TEXT_MAP[type]).toBeDefined();
                expect(TYPE_VIDEO_MAP[type]).toBeDefined();
            });
        });

        it('backend types should map to valid internal types', () => {
            Object.values(BACKEND_TYPE_MAP).forEach((internalType) => {
                expect(TYPE_IMAGE_MAP[internalType]).toBeDefined();
                expect(TYPE_TEXT_MAP[internalType]).toBeDefined();
                expect(TYPE_VIDEO_MAP[internalType]).toBeDefined();
            });
        });
    });
});
