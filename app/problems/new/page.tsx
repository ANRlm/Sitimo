import { redirect } from 'next/navigation';

export default function NewProblemPage() {
  redirect('/problems/new/edit');
}
