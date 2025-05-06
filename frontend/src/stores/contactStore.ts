import { defineStore } from 'pinia';
import apiClient from '@/services/api';
import type { Contact, ContactCreate, ContactUpdate } from '@/types/Contact';

export const useContactStore = defineStore('contact', {
  state: () => ({
    contacts: [] as Contact[],
    currentContact: null as Contact | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    getContactById: (state) => (id: number): Contact | undefined => {
      return state.contacts.find(contact => contact.contact_id === id);
    },
    contactsByCustomer: (state) => (customerId: number): Contact[] => {
      return state.contacts.filter(contact => contact.customer_id === customerId);
    },
    isLoading: (state) => state.loading,
    hasError: (state) => state.error !== null
  },

  actions: {
    /**
     * Fetch all contacts (globally)
     */
    async fetchAllContacts(search?: string) {
      try {
        this.loading = true;
        this.error = null;

        const params = search ? { search } : {};
        const response = await apiClient.get<Contact[]>('/api/contacts', { params });
        this.contacts = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch contacts';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch contacts for a specific customer
     */
    async fetchContactsByCustomer(customerId: number) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Contact[]>(
          `/api/customers/${customerId}/contacts`
        );
        this.contacts = response.data;

        return response.data;
      } catch (err) {
        this.error = `Failed to fetch contacts for customer ${customerId}`;
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch a contact by ID
     */
    async fetchContactById(contactId: number, customerId?: number) {
      try {
        this.loading = true;
        this.error = null;

        // Use the scoped API if customer ID is provided
        const url = customerId
          ? `/api/customers/${customerId}/contacts/${contactId}`
          : `/api/contacts/${contactId}`;

        const response = await apiClient.get<Contact>(url);
        this.currentContact = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch contact details';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Create a new contact for a customer
     */
    async createContact(customerId: number, contact: ContactCreate) {
      try {
        this.loading = true;
        this.error = null;

        // Ensure customer ID is set
        const contactData = {
          ...contact,
          customer_id: customerId
        };

        const response = await apiClient.post<Contact>(
          `/api/customers/${customerId}/contacts`,
          contactData
        );

        // Add new contact to the list
        this.contacts.push(response.data);

        return response.data;
      } catch (err) {
        this.error = 'Failed to create contact';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update an existing contact
     */
    async updateContact(customerId: number, contact: ContactUpdate) {
      try {
        this.loading = true;
        this.error = null;

        // Ensure customer ID is set
        const contactData = {
          ...contact,
          customer_id: customerId
        };

        const response = await apiClient.put<Contact>(
          `/api/customers/${customerId}/contacts/${contact.contact_id}`,
          contactData
        );

        // Update contact in the list
        const index = this.contacts.findIndex(
          (c) => c.contact_id === contact.contact_id
        );

        if (index !== -1) {
          this.contacts[index] = response.data;
        }

        // Update current contact if it's the one being edited
        if (this.currentContact?.contact_id === contact.contact_id) {
          this.currentContact = response.data;
        }

        return response.data;
      } catch (err) {
        this.error = 'Failed to update contact';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Delete a contact
     */
    async deleteContact(customerId: number, contactId: number) {
      try {
        this.loading = true;
        this.error = null;

        await apiClient.delete(`/api/customers/${customerId}/contacts/${contactId}`);

        // Remove contact from the list
        this.contacts = this.contacts.filter(
          (contact) => contact.contact_id !== contactId
        );

        // Reset current contact if it's the one being deleted
        if (this.currentContact?.contact_id === contactId) {
          this.currentContact = null;
        }

        return true;
      } catch (err) {
        this.error = 'Failed to delete contact';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Check if an email already exists
     */
    async checkEmailExists(email: string): Promise<boolean> {
      try {
        const response = await apiClient.get<{ exists: boolean }>('/api/contacts/check', {
          params: { email }
        });
        return response.data.exists;
      } catch (err) {
        console.error('Error checking email existence:', err);
        throw err;
      }
    }
  }
});
