# Requesting access to the OASA telematics data

leoforio is a free, open source, non-commercial app that shows Athenians when
their bus is coming. It currently reads the undocumented OASA Telematics
service.

Asking OASA for written permission is the one step no comparable project has
taken, and it is the step that would remove the only real legal ambiguity in
the project. It costs one email and no development time, and it can run in
parallel with the build.

## Who to write to

| Contact | Why |
|---|---|
| `oasa@oasa.gr` | the published contact address of OASA, listed in the GTFS feed as `agency_email` |
| O2 Hub, `data.hcap.gr` | HCAP is OASA's parent company and runs a formal open data and open API portal. OASA has contributed datasets there, so it is a channel where access requests are structurally possible. |
| Advanced Transport Telematics S.A. | the private partner in the Public-Private Partnership that built and operates the telematics system. The intellectual property in the system, and possibly in its output, may sit with this company rather than with OASA. A permission request that ignores them may be answered by OASA and still leave a second rights-holder unaddressed. |

## What to ask for

- Access to the telematics data needed for a non-commercial, open source app:
  vehicle positions, stop arrivals and timetables.
- The option of a written agreement that says what a non-commercial project
  may do, and on what terms.

## What to offer

State these commitments up front, because they are what makes the request
easy to approve:

- The project is free, open source, non-commercial, and carries no
  advertising, no analytics and no tracking.
- The data is queried live when a user opens the app, held in a cache of
  seconds, and never stored, never aggregated into a dataset and never
  resold or redistributed.
- OASA is credited as the data source, and the app states that it is
  unofficial and unaffiliated.
- The project will follow any conditions OASA sets, including attribution
  wording and rate limits.
- The backend can be pointed at a server in Greece, so the traffic pattern is
  comparable to a normal user of the official service.

## English draft

Subject: Request for access to OASA telematics data for a non-commercial open source app

Dear OASA team,

I am the developer of leoforio, a free, open source app that helps people in
Athens find out when their bus is actually coming. The project is not
commercial: it has no advertising, no analytics, no tracking and no paid
features, and the source code is public at
https://github.com/ksohos333/leoforio

At the moment the app reads the telematics service for vehicle positions,
arrivals and timetables. Because that service is undocumented, I would
rather ask than assume: is it possible for a non-commercial, open source
project like this one to use that data, and would OASA be willing to confirm
the terms in writing?

What the project would do with the data:

- query it live when a user opens the app, and hold the response in a short
  lived cache;
- never store the data, never build a stored dataset from it and never
  redistribute it;
- credit OASA as the source, and state clearly that the app is unofficial
  and not affiliated with or endorsed by OASA;
- run its backend from Greece, and follow any rate limits or conditions you
  set.

I am happy to adapt the app to whatever conditions you require, and I am
glad to answer any questions about how the data is used.

Thank you for your time.

Kind regards,
Konstantinos Sohos
https://github.com/ksohos333/leoforio

## Greek draft

Θέμα: Αίτημα για πρόσβαση στα δεδομένα τηλεματικής ΟΑΣΑ για μη εμπορική εφαρμογή ανοιχτού κώδικα

Αγαπητοί κύριοι,

Είμαι ο δημιουργός του leoforio, μιας δωρεάν, ανοιχτού κώδικα εφαρμογής που
βοηθά τους κατοίκους της Αθήνας να μάθουν πότε έρχεται το λεωφορείο τους. Η
εφαρμογή δεν είναι εμπορική: δεν έχει διαφημίσεις, δεν έχει αναλυτικά στοιχεία,
δεν παρακολουθεί χρήστες και δεν έχει επί πληρωμή. Ο πηγαίος κώδικας είναι
δημόσιος στο https://github.com/ksohos333/leoforio

Αυτή τη στιγμή η εφαρμογή διαβάζει το σύστημα τηλεματικής για θέσεις
οχημάτων, αφίξεις και προγραμματισμούς. Επειδή το σύστημα δεν τεκμηριώνεται,
προτιμώ να ρωτήσω αντί να υποθέσω: είναι δυνατόν ένα μη εμπορικό, ανοιχτού
κώδικα project όπως αυτό να χρησιμοποιεί τα δεδομένα αυτά, και θα μπορούσε η
ΟΑΣΑ να επιβεβαιώσει τους όρους γραπτώς;

Τι θα κάνει το project με τα δεδομένα:

- θα τα ζητά σε πραγματικό χρόνο όταν ο χρήστης ανοίγει την εφαρμογή, με
  σύντομη προσωρινή αποθήκευση στη μνήμη;
- δεν θα τα αποθηκεύει, δεν θα δημιουργεί μόνιμη βάση από αυτά και δεν θα
  τα αναδιανέμει σε τρίτους;
- θα αναφέρει την ΟΑΣΑ ως πηγή των δεδομένων και θα ξεκαθαρίζει ότι η
  εφαρμογή είναι επίσημη και δε συνδέεται με την ΟΑΣΑ;
- θα εκτελεί το backend του από Ελλάδα και θα ακολουθεί όποια όρια ρυθμού
  ή όρους θα θέσετε.

Είμαι διαθέσιμος να προσαρμόσω την εφαρμογή σε όποιες απαιτήσεις θέσετε και
σας ευχαριστώ για τον χρόνο σας.

Με εκτίμηση,
Κωνσταντίνος Σόχος
https://github.com/ksohos333/leoforio

## If there is no answer

A reply of "no", or no reply at all, does not change the legal position in
any way: the project then sits exactly where the other open source projects
sit, with the additional fact that a good faith request was made and
documented. The practical mitigations are already in place, see
[legal.md](./legal.md).
