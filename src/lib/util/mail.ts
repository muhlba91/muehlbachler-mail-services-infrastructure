/**
 * Creates the mailname.
 *
 * @param {string} domain the domain
 * @returns {string} the mailname
 */
export const getMailname = (domain: string): string => `mail.${domain}`;
