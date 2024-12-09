<template>
  <div class="dropdown-menu">
    <div 
      v-for="group in groups"
      :key="group[0]"
      class="dropdown-item" 
      :class="{ 'selected': group[0] === selectedGroup[0] }"
      @click="selectGroup(group)"
    >
      <img :src="group[1]" class="option-image" />
      {{ group[0] }}
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  groups: {
    type: Array,
    required: true,
  },
  modelValue: {
    type: Array,
    required: true,
  },
});

const emit = defineEmits(['update:modelValue']);

const selectedGroup = ref([...props.modelValue]);

function selectGroup(group) {
  selectedGroup.value = group;
  emit('update:modelValue', group);
}
</script>

<style scoped>
.dropdown-menu {
  border: 1px solid #ccc;
  border-radius: 4px;
  width: 200px;
  padding: 5px;
  background-color: #fff;
}

.dropdown-item {
  padding: 8px;
  display: flex;
  align-items: center;
  cursor: pointer;
}

.dropdown-item.selected {
  background-color: #e6f7ff;
}

.option-image {
  width: 24px;
  height: 24px;
  margin-right: 8px;
}
</style>
