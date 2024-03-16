import Parchment from 'parchment';

export class MentionBlot extends Parchment.Inline {
  static blotName = 'mention';
  static tagName = 'A';

  static create(value) {
    const node = super.create();
    node.textContent = value;
    return node;
  }

  static formats(node) {
    return node.textContent || true;
  }
}

Parchment.register(MentionBlot);