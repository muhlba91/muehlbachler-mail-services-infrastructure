import { StringMap } from '../../../model/map';

const LOCATIONS_DATACENTER: StringMap<string> = {
  fsn1: 'fsn1-dc14',
  nbg1: 'nbg1-dc3',
};

/**
 * Converts a location to the datacenter identifier.
 *
 * @param {string} location the location
 * @return {string} the datacenter identifier
 */
export const locationToDatacenter = (location: string): string =>
  LOCATIONS_DATACENTER[location] || 'fsn1-dc14';
