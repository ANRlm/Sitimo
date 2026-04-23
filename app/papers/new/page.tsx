import { redirect } from 'next/navigation';

export default function NewPaperPage() {
  redirect('/papers/new/editor');
}
