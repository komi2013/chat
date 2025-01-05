<template>
  <div>
    <input
      v-model="searchTerm"
      type="text"
      placeholder="ユーザー検索"
      @input="filterResults"
    />
    <ul>
      <li v-for="item in filteredResults" :key="item.id" @click="addToSelection(item)">
        <img v-if="item.image && item.image !== 'null'" :src="item.image" alt="icon" width="50" />
        <span>{{ item.name }}</span>
      </li>
    </ul>
    <hr />
    <div v-for="item in selectedItems" :key="item.id" class="selected-item">
      <img v-if="item.image && item.image !== 'null'" :src="item.image"/>
      <span>{{ item.name }}</span>
      <button @click="removeFromSelection(item)" class="remove-btn">x</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";

const props = defineProps({
  channel: Object,
  person: Object,
});

const allAliases = props.channel.allAliases;
const groupAliases = props.channel.groupAliases;
const searchTerm = ref("");
const selectedItems = ref(props.person ? [props.person] : []);
const filteredResults = computed(() => {
  if (!searchTerm.value || typeof searchTerm.value !== "string") return [];

  const results = [
    ...allAliases.map((item) => ({
      id: `${item[0] || ""}${item[1] || ""}`,
      name: item[0] || "",
      image: item[1] || null,
    })),
    ...groupAliases.map((item) => ({
      id: item[0] || "",
      name: item[0] || "",
      image: item[1] || null,
    })),
  ].filter(
    (item) =>
      item.name.includes(searchTerm.value)
  );
  return results.filter(
    (item) => !selectedItems.value.some((selected) => selected.id === item.id)
  );
});

const emit = defineEmits(["update:selectedItems"]);

const addToSelection = (item) => {
  if (!selectedItems.value.find((selected) => selected.id === item.id)) {
    selectedItems.value.push(item);
    emitSelectedItems(1, item);
  }
};

const removeFromSelection = (item) => {
  selectedItems.value = selectedItems.value.filter(
    (selected) => selected.id !== item.id
  );
  emitSelectedItems(-1, item);
};

// const handleSelectedItemsChange = (change) => {
//   const { diff, item } = change;

//   if (diff === 1) {
//     // console.log("Item added:", item);
//   } else if (diff === -1) {
//     // console.log("Item removed:", item);
//   }
// };

const emitSelectedItems = (diff, item) => {
  emit("update:selectedItems", {
    diff,
    item,
  });
};

</script>

<style scoped>
ul {
  list-style-type: none;
  padding: 0;
}

li {
  cursor: pointer;
  display: flex;
  align-items: center;
  margin-bottom: 5px;
}

.selected-item {
  display: inline-block;
  border: 1px solid #3498db;
  margin: 5px 10px 5px 0;
  border-radius: 5px;
  padding: 3px;
}

img {
  margin-right: 10px;
  max-width: 20px;
  max-height: 20px;
}
</style>
