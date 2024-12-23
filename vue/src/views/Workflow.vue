<script setup>
import { ref, onMounted } from 'vue';

const tickets = ref([]);

// Function to open IndexedDB and fetch all tickets
async function fetchTickets() {
  try {
    tickets.value = await getAllIDBs('ticket'); // Assuming 'tickets' is the table name
  } catch (error) {
    console.error('Error fetching tickets:', error);
  }
}

// Fetch tickets when the component is mounted
onMounted(() => {
  fetchTickets();
});

// Placeholder openDatabase function - replace this with your actual implementation
async function openDatabase() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('workflowDB', 1);

    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };

    request.onsuccess = (event) => {
      const db = event.target.result;
      resolve(db);
    };

    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      db.createObjectStore('tickets', { keyPath: 'ticketID' });
    };
  });
}
</script>

<template>
  <div>
    <h1>チケット一覧</h1>
    
    <!-- Link to the workflow page -->
    <a href="/ticket/">作成</a>

    <!-- Display list of tickets -->
    <ul>
      <li v-for="ticket in tickets" :key="ticket.ticketID">
        <!-- Make the ticket title clickable -->
        <a :href="'/ticket/' + ticket.ticketID">{{ ticket.title }}</a>
      </li>
    </ul>

    <!-- Show a message if there are no tickets -->
    <p v-if="!tickets.length">チケットがありません</p>
  </div>
</template>

<style scoped>
/* Add some basic styling */
ul {
  list-style-type: none;
  padding: 0;
}

li {
  margin-bottom: 10px;
}

a {
  color: #007bff;
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}
</style>
