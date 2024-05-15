- We are gonna build a task queue library

- We will be taking inspiration from https://github.com/hibiken/asynq

# Requirements

- User can submit a task
- USer can view the status of the task
- User can specify the repetitive configurations (cron)
- User can specify the priority of the task

# High level design (HLD)

The HLD focuses on the system architecture and the interaction between the components. Here is the simplified overview of the job scheduler

1. Components

   - **Client Api** :- Interface for submitting job to the queue. Jobs can have various priorities , retries and schedules
   - **Job queue** :- Persistent storage (like redis) to store job data , ensuring durability during restarts
   - **WorkerPool** :- A set of workers that execute jobs from the queue. The system should manage workers concurrently according to the workload.
   - **Scheduler** :- manages time based job scheduling , ensures that scheduled jobs are enqueued at the correct time.
   - **Result store** :- For storing job results or states , which can be queried by the clients.

2. Workflow

   - jobs are submitted to the queue via the client api
   - The scheduler monitors for time based jobs and enqueues them accordingly
   - Workers continuously poll the queue for new jobs and executes them.
   - Job execution results can be stored in the Result store for later retrieval.

3. Scalability and Reliability

   - Use a distributed queue system (i.e Redis) for scalability and reliability
   - Implement health checks for the automatic recovery of worker failures.
   - Include rate limiting and backoff strategies for job retries

4. Security
   - Secure the client api with authentication and authorization
   - Ensure data at rest and in transit is encrypted if sensitive information is handled

# Low level design

The LLD involves more detailed planning of the system's component including data models , API's and the internal logic of each component

1. Data models :-

   - Define job structures , including field for the ID , type , payload , priority ,status , retries and schedule

2. API design :-

   - Design a restful or gRPC API's for job submission , cancellation and status queries

3. Worker Implementation :-

   - Implement workers at go routines with a mechanism of dynamically adjust the number of concurrent workers based on load.
   - Use channels for worker job assignment and to handle job results or failure

4. Scheduler logic :-

   - Implement time based scheduling using a time wheel or similar algorithm for efficient job scheduling.

5. Queue management :-
   - Choose a persistent queue system like redis and design the job serialization and deserialization logic
   - Implement priority queue logic if needed.

# Project structure

```
bashCopy code
/job-scheduler
    /cmd
        /scheduler    # Main application
    /pkg
        /api          # Client api definition and implementation
        /worker       # Worker logic and lifecycle component
        /queue        # Queue management including interactions with redis
        /scheduler    # scheduler logic with time based scheduling
        /store        # Result store logic
    /internal
        /configuration # configuration management
        /util          # utility functions and common helpers
    /scrips           # utility scrips eg for setup and maintenance
```
