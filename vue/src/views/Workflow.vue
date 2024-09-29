<template>
  <div class="ticket-edit-page" v-if="ticket">
    <h1>Ticket #{{ ticket.id }}</h1>

    <div class="ticket-form">
      <label>Status</label>
      <select v-model="editableTicket.status">
        <option v-for="(statusText, index) in statusOptions" :key="index" :value="index">
          {{ statusText }}
        </option>
      </select>

      <label>Title</label>
      <input v-model="editableTicket.title" type="text" />

      <label>Assignee</label>
      <select v-model="editableTicket.assignee">
        <option v-for="(assigneeText, index) in assigneeOptions" :key="index" :value="index">
          {{ assigneeText }}
        </option>
      </select>

      <label>Description</label>
      <textarea v-model="editableTicket.description"></textarea>

      <button @click="saveChanges">Save Changes</button>
    </div>
  </div>
  <div v-else>
    <p>Loading ticket data...</p>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';

// Props definition
const props = defineProps({
  ticketID: String // Assuming the ticket ID is passed as a prop
});

const ticket = ref(null); // Will hold the ticket data after fetching
const editableTicket = reactive({}); // Editable copy of the ticket data

// Status and Assignee options
const statusOptions = ref(['Open', 'In Progress', 'Closed', 'Resolved']);
const assigneeOptions = ref(['John Doe', 'Jane Smith', 'Mike Johnson', 'Alice Green']);

// Function to fetch ticket data using fetch and ticketID
async function fetchTicket() {
  const fd = new FormData();
  fd.append('ticketID', props.ticketID);

  const request = new Request('/TicketGet/', {
    method: 'POST',
    body: fd,
  });

  try {
    const response = await fetch(request);

    if (response.ok) {
      const ticketData = await response.json();
      ticket.value = ticketData; // Assign the fetched data to ticket
      // Make a reactive copy for editing
      Object.assign(editableTicket, ticketData);
    } else {
      console.error('Failed to fetch ticket data', response.status);
    }
  } catch (error) {
    console.error('Error fetching ticket data:', error);
  }
}

// Method to save changes
function saveChanges() {
  console.log('Saving changes:', editableTicket);
  // Here you would submit the updated ticket data via an API request
  // Example:
  // axios.post('/api/ticketUpdate', editableTicket)
  //   .then(response => console.log('Ticket updated successfully', response))
  //   .catch(error => console.error('Error updating ticket', error));
}

// Fetch ticket data when the component is mounted
onMounted(() => {
  fetchTicket();
});
</script>

<style scoped>
.ticket-edit-page {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.ticket-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

label {
  font-weight: bold;
}

input, select, textarea {
  padding: 8px;
  font-size: 14px;
  border-radius: 4px;
  border: 1px solid #ccc;
}

textarea {
  resize: vertical;
  height: 100px;
}

button {
  padding: 10px 15px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}
</style>
