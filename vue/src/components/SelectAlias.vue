<template>
  <div>
    <input
      v-if="editable"
      v-model="searchTerm"
      type="text"
      :placeholder="placeholder || 'ユーザー検索'"
      @input="filterResults"
    /><br>
    <ul v-if="filteredResults.length > 0" class="dropdown-menu">
      <li v-for="item in filteredResults" :key="item.id" @click="addToSelection(item)">
        <img v-if="item.image && item.image.charAt(0) != ','" 
          :src="item.image" class="min-icon">
        <span v-if="item.image && item.image.charAt(0) == ','"
          class="min-icon" 
          :style="'background-color:' + item.image.split(',')[2] ">
            <span>{{item.image.split(',')[1]}}</span>
        </span>

        <span>{{ item.name }}</span>
      </li>
    </ul>
    <div v-for="item in selectedAlias" :key="item.id" class="selected-item">
      <img v-if="item.image && item.image.charAt(0) != ','" 
        :src="item.image" class="min-icon">
      <span v-if="item.image && item.image.charAt(0) == ','"
        class="min-icon" 
        :style="'background-color:' + item.image.split(',')[2] ">
          <span>{{item.image.split(',')[1]}}</span>
      </span>
      <span>{{ item.name }}</span>
      <button v-if="editable" @click="removeFromSelection(item)" class="remove-btn">x</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";

const props = defineProps({
  aliases: Object,
  modelValue: Array,
  editable: Boolean,
  placeholder: String
});

console.log('props.modelValue', props.modelValue);
console.log('props.aliases', props.aliases);

const emit = defineEmits(["update:modelValue"]);
const searchTerm = ref("");
const selectedAlias = ref([]);
if (Array.isArray(props.modelValue)) {
  selectedAlias.value = props.modelValue.map((name) => {
    const match = props.aliases.find((alias) => alias.aliasName === name);
    return match
      ? { id: match.aliasID, name: match.aliasName, image: match.aliasImg }
      : null;
  }).filter(Boolean);
}

console.log('selectedAlias.value', selectedAlias.value);


const filteredResults = computed(() => {
  if (!searchTerm.value || typeof searchTerm.value !== "string") return [];
  const results = [
    ...props.aliases.map((d) => ({
      id: d.aliasID,
      name: d.aliasName,
      image: d.aliasImg,
    })),
  ].filter(
    (d) =>
      d.name.includes(searchTerm.value)
  );
  return results.filter(
    (d) => !selectedAlias.value.some((selected) => selected.id === d.id)
  );
});

const addToSelection = (item) => {
  if (!selectedAlias.value.find((selected) => selected.id === item.id)) {
    selectedAlias.value.push(item);
    emitNames();
  	searchTerm.value = "";
  }
};

const removeFromSelection = (item) => {
  selectedAlias.value = selectedAlias.value.filter(
    (selected) => selected.id !== item.id
  );
  emitNames();
};

const emitNames = () => {
  const names = selectedAlias.value.map(item => item.name);
  emit("update:modelValue", names);
};

</script>

<style scoped>

.dropdown-menu {
  position: absolute;
  left: 0;
  background-color: #fff;
  border: 1px solid #ccc;
  border-radius: 4px;
  z-index: 20;
  max-height: 200px;
  overflow-y: auto;
  list-style: none;
  padding: 0;
  margin: 0;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dropdown-menu li {
  padding: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.dropdown-menu li:hover {
  background-color: #f0f0f0;
}

.selected-item {
  display: inline-block;
  border: 1px solid #3498db;
  margin: 5px 10px 5px 0;
  border-radius: 5px;
  padding: 3px;
}

/*.user-search {
  margin-top: 10px;
  padding: 10px;
}*/

.min-icon {
  width: 26px;
  max-width: 26px;
  height: 26px;
  max-height: 26px;
  border-radius: 4px;
  display: inline-flex;
  vertical-align: middle;
  justify-content: center;
  align-items: center;
}

</style>
