class CreateRuniqJobs < ActiveRecord::Migration[8.1]
  def change
    create_table :runiq_jobs do |t|
      t.string :queue_name, null: false
      t.string :job_type, null: false
      t.binary :payload, null: false
      t.string :status, null: false, default: 'pending'
      t.integer :attempts, null: false, default: 0
      t.datetime :run_at, null: false, default: -> { 'CURRENT_TIMESTAMP' }
      t.datetime :locked_at
      t.timestamps
    end

    # Índice parcial otimizado para o worker achar as tarefas pendentes instantaneamente
    add_index :runiq_jobs, [:queue_name, :status, :run_at], 
              name: 'idx_runiq_jobs_pending', 
              where: "status = 'pending'"
  end
end
